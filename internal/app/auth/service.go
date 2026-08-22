package auth

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"agropal/internal/infra/queue"
	"agropal/pkg"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	verificationCodeTTL        = 10 * time.Minute
	verificationResendCooldown = 60 * time.Second
	verificationMaxAttempts    = 5
)

type verificationPurpose string

const (
	verificationPurposeRegister      verificationPurpose = "register"
	verificationPurposeLogin         verificationPurpose = "login"
	verificationPurposeResetPassword verificationPurpose = "reset_password"
)

type verificationStatus string

const (
	verificationStatusPending  verificationStatus = "pending"
	verificationStatusVerified verificationStatus = "verified"
)

type Service struct {
	logger     *pkg.Logger
	transactor database.Transactor
	generator  *identity.Generator
	jwt        dAuth.TokenProvider
	redis      *redis.Client
	queue      queue.Publisher
	authRepo   dAuth.Repository
	userRepo   dUser.Repository
}

func NewService(
	logger *pkg.Logger,
	t database.Transactor,
	g *identity.Generator,
	jwt dAuth.TokenProvider,
	redis *redis.Client,
	mq queue.Publisher,
	ar dAuth.Repository,
	ur dUser.Repository,
) *Service {
	return &Service{
		logger:     logger,
		transactor: t,
		generator:  g,
		jwt:        jwt,
		redis:      redis,
		queue:      mq,
		authRepo:   ar,
		userRepo:   ur,
	}
}

/* ==================== Email ====================== */

// verificationRecord is stored in Redis in place of the raw code so that a
// compromised cache (or an operator with redis-cli access) cannot read out
// active codes and impersonate a user mid-flow.
type verificationRecord struct {
	CodeHash  string             `json:"code_hash"`
	Status    verificationStatus `json:"status"`
	ExpiresAt time.Time          `json:"expires_at"`
	Attempts  int                `json:"attempts"`
}

func hashVerificationCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func (s *Service) SendVerificationCode(ctx context.Context, purpose verificationPurpose, email string) error {
	redisKey := fmt.Sprintf("%s:%s", string(purpose), email)

	// Refuse to reissue a code while a recent one is still fresh, so a caller
	// can't be used to spam a victim's inbox or hammer the cache.
	if ttl, err := s.redis.TTL(ctx, redisKey).Result(); err == nil && ttl > verificationCodeTTL-verificationResendCooldown {
		return dAuth.ErrRateLimited
	}

	verificationCode, err := s.generator.GenerateHexID()
	if err != nil {
		return err
	}

	record := verificationRecord{
		CodeHash:  hashVerificationCode(verificationCode),
		Status:    verificationStatusPending,
		ExpiresAt: time.Now().UTC().Add(verificationCodeTTL),
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if err = s.redis.Set(ctx, redisKey, payload, verificationCodeTTL).Err(); err != nil {
		return err
	}

	// The plaintext code only ever leaves this function on the queue message
	// below; it is never logged and Redis only holds its hash.
	task := queue.EmailVerificationTask{
		Purpose:   string(purpose),
		To:        email,
		Code:      verificationCode,
		ExpiresAt: record.ExpiresAt,
	}
	taskPayload, err := json.Marshal(task)
	if err != nil {
		return err
	}

	if err = s.queue.PublishMessage(ctx, taskPayload); err != nil {
		s.logger.Error(ctx, shared.System, fmt.Sprintf("Failed to enqueue verification email for %s: %s", email, err.Error()))
		// Don't leave a code the user can never receive sitting in the cache.
		_ = s.redis.Del(ctx, redisKey).Err()
		return err
	}

	return nil
}

func (s *Service) VerifyCode(ctx context.Context, purpose verificationPurpose, email, token string) error {
	redisKey := fmt.Sprintf("%s:%s", string(purpose), email)
	redisValue, err := s.redis.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return dAuth.ErrVerificationCodeNotFound
	} else if err != nil {
		s.logger.Error(ctx, shared.System, fmt.Sprintf("redis key(%s) - error: %s", redisKey, err.Error()))
		return dAuth.ErrVerificationCodeNotFound
	}

	var record verificationRecord
	if err = json.Unmarshal([]byte(redisValue), &record); err != nil {
		return dAuth.ErrVerificationCodeInvalid
	}

	if record.Status == verificationStatusVerified {
		return dAuth.ErrVerificationCodeUsed
	}
	if time.Now().UTC().After(record.ExpiresAt) {
		_ = s.redis.Del(ctx, redisKey).Err()
		return dAuth.ErrVerificationCodeExpired
	}
	if record.Attempts >= verificationMaxAttempts {
		_ = s.redis.Del(ctx, redisKey).Err()
		return dAuth.ErrTooManyAttempts
	}

	// Constant-time comparison of the hashes avoids leaking how many
	// leading bytes of the guess were correct via response timing.
	tokenHash := hashVerificationCode(token)
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(record.CodeHash)) != 1 {
		record.Attempts++
		if payload, mErr := json.Marshal(record); mErr == nil {
			ttl := time.Until(record.ExpiresAt)
			if ttl > 0 {
				_ = s.redis.Set(ctx, redisKey, payload, ttl).Err()
			}
		}
		return dAuth.ErrVerificationCodeInvalid
	}

	record.Status = verificationStatusVerified
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ttl := time.Until(record.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	if err = s.redis.Set(ctx, redisKey, payload, ttl).Err(); err != nil {
		return err
	}

	return nil
}

/* ==================== Identity ====================== */

// Register creates the account in an unverified state and emails a
// verification code; the account cannot log in until VerifyRegistration
// confirms it.
func (s *Service) Register(ctx context.Context, cmd CreateUserCommand) error {
	exists, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), cmd.Email)
	if err != nil {
		return err
	}
	if exists.ID != uuid.Nil {
		return dAuth.ErrAlreadyRegistered
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	cmd.Password = string(hashedPassword)

	userID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	user := cmd.ToDomain(userID)

	// Send the verification code before creating the account: if this fails,
	// nothing was persisted and the caller can just retry. Creating the user
	// first would risk a stuck unverified account with no code ever sent if
	// this step failed afterward.
	if err = s.SendVerificationCode(ctx, verificationPurposeRegister, cmd.Email); err != nil {
		return err
	}

	if err = s.userRepo.CreateUser(ctx, s.transactor.DB(), user); err != nil {
		return err
	}

	return nil
}

// VerifyRegistration confirms a pending registration's code and marks the
// account's email as verified, unblocking Login.
func (s *Service) VerifyRegistration(ctx context.Context, email, code string) (dAuth.TokenPair, error) {
	if err := s.VerifyCode(ctx, verificationPurposeRegister, email, code); err != nil {
		return dAuth.TokenPair{}, err
	}

	user, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), email)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	if err = s.userRepo.VerifyUserEmail(ctx, s.transactor.DB(), user.ID, time.Now().UTC()); err != nil {
		return dAuth.TokenPair{}, err
	}

	claims := shared.ClaimInfo{
		UserID: user.ID,
		Email:  user.Account.Email,
		Role:   string(user.Security.Role),
	}

	tokenPair, err := s.jwt.GenerateTokenPair(ctx, claims)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	if err = s.authRepo.CreateRefreshToken(ctx, s.transactor.DB(), tokenPair.RefreshToken); err != nil {
		return dAuth.TokenPair{}, err
	}

	return tokenPair, nil
}

// Login validates the account's credentials and, on success, emails a
// one-time code rather than issuing tokens directly; VerifyLogin completes
// the sign-in once that code is confirmed.
func (s *Service) Login(ctx context.Context, cmd LoginUserCommand) error {
	var user dUser.User
	var err error

	if cmd.Email != nil {
		user, err = s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), *cmd.Email)
		if err != nil {
			return err
		}
	}
	if cmd.Username != nil {
		user, err = s.userRepo.FindUserByUsername(ctx, s.transactor.DB(), *cmd.Username)
		if err != nil {
			return err
		}
	}

	if user == (dUser.User{}) {
		return dAuth.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Security.Password), []byte(cmd.Password)); err != nil {
		return dAuth.ErrInvalidPassword
	}

	if user.Metadata.EmailVerifiedAt == nil {
		return dAuth.ErrEmailNotVerified
	}

	return s.SendVerificationCode(ctx, verificationPurposeLogin, user.Account.Email)
}

// VerifyLogin confirms the one-time code from Login and, if valid, issues
// the token pair.
func (s *Service) VerifyLogin(ctx context.Context, email, code string) (dAuth.TokenPair, error) {
	if err := s.VerifyCode(ctx, verificationPurposeLogin, email, code); err != nil {
		return dAuth.TokenPair{}, err
	}

	user, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), email)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	claims := shared.ClaimInfo{
		UserID: user.ID,
		Email:  user.Account.Email,
		Role:   string(user.Security.Role),
	}

	tokenPair, err := s.jwt.GenerateTokenPair(ctx, claims)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	if err = s.authRepo.CreateRefreshToken(ctx, s.transactor.DB(), tokenPair.RefreshToken); err != nil {
		return dAuth.TokenPair{}, err
	}

	return tokenPair, nil
}

/* ==================== Tokens ====================== */

func (s *Service) Refresh(ctx context.Context, refreshID string) (dAuth.TokenPair, error) {
	claims, err := s.jwt.ValidateToken(ctx, shared.TypeRefresh, refreshID)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	storedToken, err := s.authRepo.GetRefreshToken(ctx, s.transactor.DB(), refreshID)
	if err != nil {
		return dAuth.TokenPair{}, dAuth.ErrTokenRevoked
	}

	if claims.UserID != storedToken.UserID {
		return dAuth.TokenPair{}, dAuth.ErrTokenInvalid
	}

	user, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), claims.UserID)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	newClaims := shared.ClaimInfo{
		UserID: user.ID,
		Email:  user.Account.Email,
		Role:   string(user.Security.Role),
	}

	newTokenPair, err := s.jwt.GenerateTokenPair(ctx, newClaims)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	err = s.transactor.WithTx(ctx, func(q database.Querier) error {
		if err = s.authRepo.DeleteRefreshToken(ctx, q, refreshID); err != nil {
			return err
		}
		if err = s.authRepo.CreateRefreshToken(ctx, q, newTokenPair.RefreshToken); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	return newTokenPair, nil
}

func (s *Service) Invalidate(ctx context.Context, refreshID string) error {
	claims, err := s.jwt.ValidateToken(ctx, shared.TypeRefresh, refreshID)
	if err != nil {
		return err
	}

	storedToken, err := s.authRepo.GetRefreshToken(ctx, s.transactor.DB(), refreshID)
	if err != nil {
		return dAuth.ErrTokenRevoked
	}

	if claims.UserID != storedToken.UserID {
		return dAuth.ErrTokenInvalid
	}

	user, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), claims.UserID)
	if err != nil {
		return err
	}
	if user == (dUser.User{}) {
		return dAuth.ErrTokenInvalid
	}

	if err = s.authRepo.DeleteRefreshToken(ctx, s.transactor.DB(), refreshID); err != nil {
		return err
	}

	return nil
}

/* ==================== Recovery ====================== */

// RequestPasswordReset emails a one-time code for the given email. It
// silently succeeds when the email is unknown, to avoid leaking account
// existence to the caller.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	_, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), email)
	if err != nil {
		if errors.Is(err, dUser.ErrNotFound) {
			return nil
		}
		return err
	}

	return s.SendVerificationCode(ctx, verificationPurposeResetPassword, email)
}

// ConfirmPasswordReset verifies the one-time code from RequestPasswordReset
// and, if valid, sets the account's new password.
func (s *Service) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	if err := s.VerifyCode(ctx, verificationPurposeResetPassword, email, code); err != nil {
		return err
	}

	user, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), email)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateUserPassword(ctx, s.transactor.DB(), user.ID, string(hashedPassword))
}

/* ==================== Authorization ====================== */

func (s *Service) Authorize(ctx context.Context, userID uuid.UUID, requiredRole string) error {
	user, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), userID)
	if err != nil {
		return err
	}

	if user.Security.Role == config.RoleAdmin {
		return nil
	}

	if requiredRole != "" && string(user.Security.Role) != requiredRole {
		return shared.ErrForbidden
	}

	return nil
}

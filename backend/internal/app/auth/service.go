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

type VerificationPurpose string

const (
	VerificationPurposeRegister      VerificationPurpose = "register"
	VerificationPurposeLogin         VerificationPurpose = "login"
	VerificationPurposeResetPassword VerificationPurpose = "reset_password"
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

func (s *Service) SendVerificationCode(ctx context.Context, purpose VerificationPurpose, email string) error {
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

// checkCode validates purpose/email/token against Redis without marking the
// code consumed. Most callers should use VerifyCode instead; checkCode
// exists for a caller (VerifyRegistration) that still has downstream work
// to do after a correct code, and needs to be able to retry that work with
// the same code if it fails, instead of burning the code on a guess that
// was actually right. A wrong guess still counts against
// verificationMaxAttempts either way.
func (s *Service) checkCode(ctx context.Context, purpose VerificationPurpose, email, token string) (verificationRecord, string, error) {
	redisKey := fmt.Sprintf("%s:%s", string(purpose), email)
	redisValue, err := s.redis.Get(ctx, redisKey).Result()
	if errors.Is(err, redis.Nil) {
		return verificationRecord{}, "", dAuth.ErrVerificationCodeNotFound
	} else if err != nil {
		s.logger.Error(ctx, shared.System, fmt.Sprintf("redis key(%s) - error: %s", redisKey, err.Error()))
		return verificationRecord{}, "", dAuth.ErrVerificationCodeNotFound
	}

	var record verificationRecord
	if err = json.Unmarshal([]byte(redisValue), &record); err != nil {
		return verificationRecord{}, "", dAuth.ErrVerificationCodeInvalid
	}

	if record.Status == verificationStatusVerified {
		return verificationRecord{}, "", dAuth.ErrVerificationCodeUsed
	}
	if time.Now().UTC().After(record.ExpiresAt) {
		_ = s.redis.Del(ctx, redisKey).Err()
		return verificationRecord{}, "", dAuth.ErrVerificationCodeExpired
	}
	if record.Attempts >= verificationMaxAttempts {
		_ = s.redis.Del(ctx, redisKey).Err()
		return verificationRecord{}, "", dAuth.ErrTooManyAttempts
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
		return verificationRecord{}, "", dAuth.ErrVerificationCodeInvalid
	}

	return record, redisKey, nil
}

// markCodeVerified flips a checked-out record to verified so it can't be
// replayed, preserving its remaining TTL.
func (s *Service) markCodeVerified(ctx context.Context, redisKey string, record verificationRecord) error {
	record.Status = verificationStatusVerified
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	ttl := time.Until(record.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	return s.redis.Set(ctx, redisKey, payload, ttl).Err()
}

func (s *Service) VerifyCode(ctx context.Context, purpose VerificationPurpose, email, token string) error {
	record, redisKey, err := s.checkCode(ctx, purpose, email, token)
	if err != nil {
		return err
	}
	return s.markCodeVerified(ctx, redisKey, record)
}

/* ==================== Identity ====================== */

// registerPendingKey namespaces the Redis key holding a not-yet-created
// account's data while it waits on VerifyRegistration, so it can't collide
// with any other feature that happens to key Redis by a bare email address.
func registerPendingKey(email string) string {
	return fmt.Sprintf("register:pending:%s", email)
}

// Register stages the account (password hashed, not yet persisted) in
// Redis and emails a verification code; VerifyRegistration only creates the
// row in Postgres once that code is confirmed, so no unverified account
// ever sits in the database. If SendVerificationCode fails, nothing beyond
// the staged Redis entry exists yet and the caller can just retry.
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

	payloadJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	if err = s.redis.Set(ctx, registerPendingKey(cmd.Email), payloadJSON, verificationCodeTTL).Err(); err != nil {
		return err
	}

	if err = s.SendVerificationCode(ctx, VerificationPurposeRegister, cmd.Email); err != nil {
		return err
	}

	return nil
}

// VerifyRegistration confirms a pending registration's code and creates the
// account, already marked verified. The code is checked but deliberately
// not consumed until CreateUser succeeds: if account creation fails (DB
// down, or a losing race against another registration for the same
// email/username), the code is still valid and the caller can retry
// VerifyRegistration without going back through Register for a new one.
func (s *Service) VerifyRegistration(ctx context.Context, email, code string) (dAuth.TokenPair, error) {
	record, redisKey, err := s.checkCode(ctx, VerificationPurposeRegister, email, code)
	if err != nil {
		return dAuth.TokenPair{}, err
	}

	pendingKey := registerPendingKey(email)
	userBytes, err := s.redis.Get(ctx, pendingKey).Bytes()
	if err != nil {
		return dAuth.TokenPair{}, err
	}
	var user dUser.User
	if err = json.Unmarshal(userBytes, &user); err != nil {
		return dAuth.TokenPair{}, err
	}

	if err = s.userRepo.CreateUser(ctx, s.transactor.DB(), user); err != nil {
		return dAuth.TokenPair{}, err
	}

	if err = s.userRepo.VerifyUserEmail(ctx, s.transactor.DB(), user.ID, time.Now().UTC()); err != nil {
		return dAuth.TokenPair{}, err
	}

	// The account now exists (and is verified) in Postgres, which is what
	// actually stops this code from being replayed into a second account -
	// a second CreateUser attempt fails on the unique email/username
	// indexes. Marking the code used and dropping the staged payload here
	// is best-effort cleanup, not load-bearing, so a failure doesn't block
	// the response.
	if err := s.markCodeVerified(ctx, redisKey, record); err != nil {
		s.logger.Error(ctx, shared.System, fmt.Sprintf("failed to mark registration code used for %s: %s", email, err.Error()))
	}
	_ = s.redis.Del(ctx, pendingKey).Err()

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

	return s.SendVerificationCode(ctx, VerificationPurposeLogin, user.Account.Email)
}

// VerifyLogin confirms the one-time code from Login and, if valid, issues
// the token pair.
func (s *Service) VerifyLogin(ctx context.Context, email, code string) (dAuth.TokenPair, error) {
	if err := s.VerifyCode(ctx, VerificationPurposeLogin, email, code); err != nil {
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

	return s.SendVerificationCode(ctx, VerificationPurposeResetPassword, email)
}

// ConfirmPasswordReset verifies the one-time code from RequestPasswordReset
// and, if valid, sets the account's new password.
func (s *Service) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	if err := s.VerifyCode(ctx, VerificationPurposeResetPassword, email, code); err != nil {
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

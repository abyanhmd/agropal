package auth

import (
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const passwordResetTokenTTL = 15 * time.Minute

type Service struct {
	transactor database.Transactor
	generator  *identity.Generator
	jwt        dAuth.TokenProvider
	authRepo   dAuth.Repository
	userRepo   dUser.Repository
}

func NewService(
	t database.Transactor,
	g *identity.Generator,
	jwt dAuth.TokenProvider,
	ar dAuth.Repository,
	ur dUser.Repository,
) *Service {
	return &Service{
		transactor: t,
		generator:  g,
		jwt:        jwt,
		authRepo:   ar,
		userRepo:   ur,
	}
}

/* ==================== Identity ====================== */

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

	return s.userRepo.CreateUser(ctx, s.transactor.DB(), user)
}

func (s *Service) Login(ctx context.Context, cmd LoginUserCommand) (dAuth.TokenPair, error) {
	var user dUser.User
	var err error

	if cmd.Email != nil {
		user, err = s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), *cmd.Email)
		if err != nil {
			return dAuth.TokenPair{}, err
		}
	}
	if cmd.Username != nil {
		user, err = s.userRepo.FindUserByUsername(ctx, s.transactor.DB(), *cmd.Username)
		if err != nil {
			return dAuth.TokenPair{}, err
		}
	}

	password := cmd.Password
	if user == (dUser.User{}) {
		return dAuth.TokenPair{}, dAuth.ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Security.Password), []byte(password)); err != nil {
		return dAuth.TokenPair{}, dAuth.ErrInvalidPassword
	}

	// Generate token pair
	claims := shared.ClaimInfo{
		UserID: user.ID,
		Email:  user.Account.Email,
		Role:   string(user.Security.Role),
	}

	// Generate an access token
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

// RequestPasswordReset issues a single-use reset token for the given email,
// stored via the refresh token repository (ExpiresAt enforces its lifetime).
// It returns an empty token without error when the email is unknown, to
// avoid leaking account existence to the caller.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, s.transactor.DB(), email)
	if err != nil {
		if errors.Is(err, dUser.ErrNotFound) {
			return "", nil
		}
		return "", err
	}

	resetToken, err := s.generator.GenerateHexID()
	if err != nil {
		return "", err
	}

	token := dAuth.RefreshToken{
		ID:        resetToken,
		UserAgent: shared.GetMetadata(ctx, shared.KeyUserAgent),
		IPAddress: shared.GetMetadata(ctx, shared.KeyIPAddress),
		ExpiresAt: time.Now().UTC().Add(passwordResetTokenTTL).Unix(),
		UserID:    user.ID,
	}

	if err = s.authRepo.CreateRefreshToken(ctx, s.transactor.DB(), token); err != nil {
		return "", err
	}

	return resetToken, nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		storedToken, err := s.authRepo.GetRefreshToken(ctx, q, token)
		if err != nil {
			return dAuth.ErrPasswordResetInvalid
		}

		if time.Now().UTC().Unix() > storedToken.ExpiresAt {
			_ = s.authRepo.DeleteRefreshToken(ctx, q, token)
			return dAuth.ErrPasswordResetExpired
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		if err = s.userRepo.UpdateUserPassword(ctx, q, storedToken.UserID, string(hashedPassword)); err != nil {
			return err
		}

		return s.authRepo.DeleteRefreshToken(ctx, q, token)
	})
}

/* ==================== Authorization ====================== */

func (s *Service) Authorize(ctx context.Context, userID uuid.UUID, requiredRole string) error {
	user, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), userID)
	if err != nil {
		return err
	}

	if user.Security.Role == dAuth.RoleAdmin {
		return nil
	}

	if requiredRole != "" && string(user.Security.Role) != requiredRole {
		return shared.ErrForbidden
	}

	return nil
}

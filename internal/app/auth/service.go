package auth

import (
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	return nil
}

/* ==================== Recovery ====================== */

func (s *Service) Authorize(ctx context.Context, userID uuid.UUID, requiredRole string) error {
	return nil
}

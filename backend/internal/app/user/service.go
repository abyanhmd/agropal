package user

import (
	"agropal/config"
	dAuth "agropal/internal/domain/auth"
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	transactor database.Transactor
	userRepo   dUser.Repository
	authRepo   dAuth.Repository
	noteRepo   dNote.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	ar dAuth.Repository,
	nr dNote.Repository,
) *Service {
	return &Service{
		transactor: t,
		userRepo:   ur,
		authRepo:   ar,
		noteRepo:   nr,
	}
}

func (s *Service) GetByID(ctx context.Context, userID uuid.UUID) (dUser.User, error) {
	user, err := s.verifyAccess(ctx, userID)
	if err != nil {
		return dUser.User{}, err
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, cmd UserCommand) error {
	var userID uuid.UUID
	if cmd.UserID == nil {
		userID = uuid.Nil
	} else {
		userID = *cmd.UserID
	}

	user, err := s.verifyAccess(ctx, userID)
	if err != nil {
		return err
	}
	cmd.UserID = &user.ID

	return s.userRepo.UpdateUser(ctx, s.transactor.DB(), cmd.ToDomain())
}

func (s *Service) UpdateUserPassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.verifyAccess(ctx, userID)
	if err != nil {
		return err
	}
	if userID == uuid.Nil {
		userID = user.ID
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Security.Password), []byte(oldPassword)); err != nil {
		return shared.ErrUnauthorized
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// A new password ends every session, and either both happen or neither does
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		if err = s.userRepo.UpdateUserPassword(ctx, q, user.ID, string(hashedPassword)); err != nil {
			return err
		}
		return s.authRepo.InvalidateSessions(ctx, q, user.ID)
	})
}

func (s *Service) Delete(ctx context.Context, deleteType shared.DeleteType, userID uuid.UUID) error {
	targetUser, err := s.verifyAccess(ctx, userID)
	if err != nil {
		return err
	}
	requesterID, _ := shared.ExtractUserID(ctx)

	nowUTC := time.Now().UTC()
	noteID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	deleteNote := dNote.Note{
		ID:        noteID,
		CreatedAt: nowUTC,
		AssetType: shared.AssetUser,
		AssetID:   targetUser.ID,
		AuthorID:  requesterID,
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		if deleteType == shared.DeleteSoft {
			deleteNote.Type = dNote.TypeArchive
			if err = s.userRepo.SoftDeleteUser(ctx, q, targetUser.ID, nowUTC); err != nil {
				return err
			}
			// A deleted account must not keep working through sessions it already has
			if err = s.authRepo.InvalidateSessions(ctx, q, targetUser.ID); err != nil {
				return err
			}
		} else if deleteType == shared.DeleteHard {
			deleteNote.Type = dNote.TypeDelete
			if err = s.userRepo.DeleteUser(ctx, q, userID); err != nil {
				return err
			}
		}
		if err = s.noteRepo.CreateNote(ctx, q, deleteNote); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) GetUsers(ctx context.Context) ([]dUser.User, error) {
	// Admin-only: this returns every user's name/email/role with no
	// scoping. Enforced here as well as at the route (PermGetManyUsers) so
	// it can't be re-exposed by a router change.
	if _, err := shared.ExtractUserID(ctx); err != nil {
		return nil, err
	}
	if shared.GetMetadata(ctx, shared.KeyRole) != string(config.RoleAdmin) {
		return nil, shared.ErrForbidden
	}

	return s.userRepo.ListUsers(ctx, s.transactor.DB())
}

/* ==================== Helper ====================== */

func (s *Service) verifyAccess(ctx context.Context, targetID uuid.UUID) (dUser.User, error) {
	requesterID, err := shared.ExtractUserID(ctx)
	if err != nil {
		return dUser.User{}, dUser.ErrInvalidUserID
	}

	// Default to self if target is Nil
	if targetID == uuid.Nil {
		targetID = requesterID
	}

	targetUser, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), targetID)
	if err != nil {
		return dUser.User{}, err
	}

	requesterRole := shared.GetMetadata(ctx, shared.KeyRole)

	isSelf := targetID == requesterID
	isAdmin := requesterRole == string(config.RoleAdmin)

	if !isSelf && !isAdmin {
		return dUser.User{}, shared.ErrForbidden
	}

	return targetUser, nil
}

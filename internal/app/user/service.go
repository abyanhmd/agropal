package user

import (
	dAuth "agropal/internal/domain/auth"
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	transactor database.Transactor
	userRepo   dUser.Repository
	noteRepo   dNote.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	nr dNote.Repository,
) *Service {
	return &Service{
		transactor: t,
		userRepo:   ur,
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

func (s *Service) UpdateUserPassword(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := s.verifyAccess(ctx, userID)
	if err != nil {
		return err
	}
	if userID == uuid.Nil {
		userID = user.ID
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	password = string(hashedPassword)

	return s.userRepo.UpdateUserPassword(ctx, s.transactor.DB(), user.ID, password)
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
		ID:          noteID,
		CreatedAt:   nowUTC,
		AssetType:   shared.AssetUser,
		AssetID:     targetUser.ID,
		AuthorID:    requesterID,
		UpdateField: "user",
		IsSystemLog: true,
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		if deleteType == shared.DeleteSoft {
			deleteNote.Type = dNote.TypeArchive
			if err = s.userRepo.SoftDeleteUser(ctx, q, userID, nowUTC); err != nil {
				return err
			}
		} else if deleteType == shared.DeleteHard {
			deleteNote.Type = dNote.TypeDelete
			if err = s.userRepo.DeleteUser(ctx, q, userID); err != nil {
				return err
			}
		}
		if err = s.noteRepo.CreateManyNote(ctx, q, []dNote.Note{deleteNote}); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) GetUsers(ctx context.Context, assetTypeStr shared.AssetType, assetID uuid.UUID) ([]dUser.User, error) {
	requesterID, _ := shared.ExtractUserID(ctx)
	_, err := s.verifyAccess(ctx, requesterID)
	if err != nil {
		return []dUser.User{}, err
	}

	var assetType shared.AssetType

	switch assetTypeStr {
	case shared.AssetUser:
		assetType = shared.AssetUser
	case shared.AssetFarm:
		assetType = shared.AssetFarm
	case shared.AssetProduct:
		assetType = shared.AssetProduct
	case shared.AssetPost:
		assetType = shared.AssetPost
	case shared.AssetComment:
		assetType = shared.AssetComment
	default:
		return nil, errors.New("invalid asset type")
	}

	return s.userRepo.ListUsersByAsset(ctx, s.transactor.DB(), assetType, assetID)
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
	isAdmin := requesterRole == string(dAuth.RoleAdmin)

	if !isSelf && !isAdmin {
		return dUser.User{}, shared.ErrForbidden
	}

	return targetUser, nil
}

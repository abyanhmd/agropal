package user

import (
	"agropal/internal/database"
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	Transactor database.Transactor
	UserRepo   dUser.Repository
}

func NewService(userRepo dUser.Repository) Service {
	return Service{
		UserRepo: userRepo,
	}
}

func (s *Service) GetByID(ctx context.Context, userID uuid.UUID) (dUser.User, error) {
	var err error

	if userID == uuid.Nil {
		userID, err = utils.ExtractUserID(ctx)
		if err != nil {
			return dUser.User{}, dUser.ErrInvalidUserID
		}
	}

	return s.UserRepo.FindByID(ctx, s.Transactor.DB(), userID)
}

func (s *Service) Update(ctx context.Context, cmd UserCommand) error {
	user := cmd.ToDomain()

	// TODO: Validate update input

	return s.UserRepo.Update(ctx, user)
}

func (s *Service) Delete(ctx context.Context, deleteType shared.DeleteType, userID uuid.UUID) error {
	var err error

	if userID == uuid.Nil {
		userID, err = utils.ExtractUserID(ctx)
		if err != nil {
			return dUser.ErrInvalidUserID
		}
	}

	nowUTC := time.Now().UTC()
	noteID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	deleteNote := dNote.Note{
		ID:        noteID,
		CreatedAt: nowUTC,
		AssetType: shared.AssetUser,
		AssetID:   userID,
	}

	switch deleteType {
	case shared.DeleteSoft:
		deleteNote.Type = dNote.NoteSoftDelete
		return s.UserRepo.SoftDelete(ctx, s.Transactor.DB(), userID, nowUTC)
	case shared.DeleteHard:
		deleteNote.Type = dNote.NoteHardDelete
		return s.UserRepo.Delete(ctx, s.Transactor.DB(), userID)
	default:
		return errors.New("invalid delete type")
	}
}

func (s *Service) GetUsers(ctx context.Context, assetTypeStr shared.AssetType, assetID uuid.UUID) ([]dUser.User, error) {
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

	return s.UserRepo.ListByAsset(ctx, assetType, assetID)
}

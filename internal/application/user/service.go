package user

import (
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
	UserRepository dUser.Repository
}

func NewService(userRepository dUser.Repository) Service {
	return Service{
		UserRepository: userRepository,
	}
}

func (s *Service) GetByID(ctx context.Context, userIDPtr *uuid.UUID) (dUser.User, error) {
	var userID uuid.UUID
	var err error

	if userIDPtr == nil {
		userID, err = utils.GetUserIDFromContext(ctx)
		if err != nil {
			return dUser.User{}, err
		}
	} else {
		userID = *userIDPtr
	}

	return s.UserRepository.GetOne(ctx, userID)
}

func (s *Service) Update(ctx context.Context, cmd UserCommand) error {
	user := cmd.ToDomain()

	// TODO: Validate update input

	return s.UserRepository.Update(ctx, user)
}

func (s *Service) Delete(ctx context.Context, deleteType shared.DeleteType, userIDPtr *uuid.UUID) error {
	var userID uuid.UUID
	var err error

	if userIDPtr == nil {
		userID, err = utils.GetUserIDFromContext(ctx)
		if err != nil {
			return err
		}
	} else {
		userID = *userIDPtr
	}

	switch deleteType {
	case shared.DeleteSoft:
		deleteNote := dNote.Note{
			NoteID:    uuid.New(),
			DateAdded: time.Now(),
			NoteType:  dNote.NoteSoftDelete,
			AssetType: shared.AssetUser,
			AssetID:   userID,
		}
		return s.UserRepository.SoftDelete(ctx, deleteNote)
	case shared.DeleteHard:
		deleteNote := dNote.Note{
			NoteID:    uuid.New(),
			DateAdded: time.Now(),
			NoteType:  dNote.NoteHardDelete,
			AssetType: shared.AssetUser,
			AssetID:   userID,
		}
		return s.UserRepository.HardDelete(ctx, userID, deleteNote)
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

	return s.UserRepository.GetByAdmin(ctx, assetType, assetID)
}

package product

import (
	"agropal/config"
	dNote "agropal/internal/domain/note"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateProduct(ctx context.Context, farmID uuid.UUID, cmd ProductCommand) (uuid.UUID, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return uuid.Nil, err
	}

	productID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	product := cmd.ToDomain(farmID, productID)

	if strings.TrimSpace(product.Unit) == "" {
		return uuid.Nil, dProduct.ErrInvalidUnit
	}

	if err = s.productRepo.CreateProduct(ctx, s.transactor.DB(), product); err != nil {
		return uuid.Nil, err
	}

	return productID, nil
}

func (s *Service) UpdateProduct(ctx context.Context, farmID, productID uuid.UUID, cmd ProductCommand) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	userID, err := shared.ExtractUserID(ctx)
	if err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		product, err := s.productRepo.FindProductByID(ctx, q, productID)
		if err != nil {
			return err
		}
		if product.FarmID != farmID {
			return shared.ErrForbidden
		}

		updateNotes := make([]dNote.Note, 0)
		updateNote := dNote.Note{
			CreatedAt:   time.Now().UTC(),
			Type:        dNote.TypeUpdate,
			AssetType:   shared.AssetProduct,
			AssetID:     productID,
			AuthorID:    userID,
			IsSystemLog: false,
		}

		if cmd.Name != nil {
			noteID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			updateNote.ID = noteID
			updateNote.UpdateField = shared.Ptr("name")
			updateNote.OldValue = shared.Ptr(product.Name)
			updateNote.NewValue = cmd.Name
			updateNotes = append(updateNotes, updateNote)

			product.Name = *cmd.Name
		}
		if cmd.Description != nil {
			noteID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			updateNote.ID = noteID
			updateNote.UpdateField = shared.Ptr("description")
			updateNote.OldValue = shared.Ptr(product.Description)
			updateNote.NewValue = cmd.Description
			updateNotes = append(updateNotes, updateNote)

			product.Description = *cmd.Description
		}
		if cmd.Image != nil {
			noteID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			updateNote.ID = noteID
			updateNote.UpdateField = shared.Ptr("image")
			updateNote.OldValue = shared.Ptr(product.Image)
			updateNote.NewValue = cmd.Image
			updateNotes = append(updateNotes, updateNote)

			product.Image = *cmd.Image
		}
		if cmd.Unit != nil {
			if strings.TrimSpace(*cmd.Unit) == "" {
				return dProduct.ErrInvalidUnit
			}
			noteID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			updateNote.ID = noteID
			updateNote.UpdateField = shared.Ptr("unit")
			updateNote.OldValue = shared.Ptr(product.Unit)
			updateNote.NewValue = cmd.Unit
			updateNotes = append(updateNotes, updateNote)

			product.Unit = *cmd.Unit
		}

		if err = s.productRepo.UpdateProduct(ctx, q, product); err != nil {
			return err
		}

		if err = s.noteRepo.CreateManyNotes(ctx, q, updateNotes); err != nil {
			return err
		}

		return nil
	})

}

func (s *Service) GetOneProduct(ctx context.Context, farmID, productID uuid.UUID) (dProduct.ProductDetail, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return dProduct.ProductDetail{}, err
	}
	productDetail, err := s.productRepo.FindProductDetailByID(ctx, s.transactor.DB(), productID)
	if err != nil {
		return dProduct.ProductDetail{}, err
	}
	if productDetail.Product.FarmID != farmID {
		return dProduct.ProductDetail{}, shared.ErrForbidden
	}

	return productDetail, nil
}

func (s *Service) GetManyProducts(ctx context.Context, farmID uuid.UUID) ([]dProduct.ProductDetail, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return []dProduct.ProductDetail{}, err
	}

	productDetails, err := s.productRepo.ListProductsByFarmID(ctx, s.transactor.DB(), farmID)
	if err != nil {
		return []dProduct.ProductDetail{}, err
	}

	return productDetails, nil
}

func (s *Service) DeleteProduct(ctx context.Context, farmID, productID uuid.UUID) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	// verifyFarm only proves the requester owns farmID; the product must also belong to it,
	// or any owner could delete another farm's product by pairing their farm ID with its ID
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		product, err := s.productRepo.FindProductByID(ctx, q, productID)
		if err != nil {
			return err
		}
		if product.FarmID != farmID {
			return shared.ErrForbidden
		}

		return s.productRepo.DeleteProduct(ctx, q, productID)
	})
}

func (s *Service) verifyFarm(ctx context.Context, farmID uuid.UUID) error {
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		userID, err := shared.ExtractUserID(ctx)
		if err != nil {
			return err
		}
		user, err := s.userRepo.FindUserByID(ctx, q, userID)
		if err != nil {
			return err
		}
		if user == (dUser.User{}) {
			return dUser.ErrInvalidUserID
		}
		farm, err := s.farmRepo.FindFarmByID(ctx, q, farmID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if farm.UserID != user.ID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		return nil
	})
}

package app

import (
	"agropal/config"
	aAuth "agropal/internal/app/auth"
	aFarm "agropal/internal/app/farm"
	aLocation "agropal/internal/app/location"
	aPost "agropal/internal/app/post"
	aProduct "agropal/internal/app/product"
	aTransaction "agropal/internal/app/transaction"
	aUser "agropal/internal/app/user"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"context"

	dAuth "agropal/internal/domain/auth"
	dFarm "agropal/internal/domain/farm"
	dLocation "agropal/internal/domain/location"
	dProduct "agropal/internal/domain/product"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"

	"agropal/internal/infra/crypto"
	"agropal/internal/infra/database"
	"agropal/internal/infra/identity"
	"agropal/internal/infra/persistence"
	"agropal/internal/infra/queue"
	"agropal/pkg"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Toolkit struct {
	Logger     *pkg.Logger
	Config     *config.Config
	Encryption *crypto.Encryption
	DB         *sqlx.DB
	Redis      *redis.Client
	Queue      *queue.RabbitMQ
}

func NewToolkit(
	logger *pkg.Logger,
	cfg *config.Config,
	enc *crypto.Encryption,
	db *sqlx.DB,
	redis *redis.Client,
	mq *queue.RabbitMQ,
) *Toolkit {
	return &Toolkit{
		Logger:     logger,
		Config:     cfg,
		Encryption: enc,
		DB:         db,
		Redis:      redis,
		Queue:      mq,
	}
}

type Container struct {
	JWT   dAuth.TokenProvider
	Queue *queue.RabbitMQ

	Auth        *AuthModule
	User        *UserModule
	Farm        *FarmModule
	Location    *LocationModule
	Product     *ProductModule
	Post        *PostModule
	Transaction *TransactionModule
}

type AuthModule struct {
	Service    *aAuth.Service
	Repository *dAuth.Repository
}
type UserModule struct {
	Service    *aUser.Service
	Repository *dUser.Repository
}
type FarmModule struct {
	Service    *aFarm.Service
	Repository *dFarm.Repository
}
type LocationModule struct {
	Service    *aLocation.Service
	Repository *dLocation.Repository
}
type ProductModule struct {
	Service    *aProduct.Service
	Repository *dProduct.Repository
}
type PostModule struct {
	Service    *aPost.Service
	Repository *dPost.Repository
}
type TransactionModule struct {
	Service    *aTransaction.Service
	Repository *dTransaction.Repository
}

func NewContainer(
	t *Toolkit,
) *Container {
	transactor := database.NewTransactor(t.DB)
	generator := identity.NewGenerator()
	jwt := crypto.NewJWTProvider(t.Logger, t.Config, t.Encryption, generator)

	/* Repository */

	authRepo := persistence.NewAuthRepository(t.Logger)
	userRepo := persistence.NewUserRepository(t.Logger)
	noteRepo := persistence.NewNoteRepository(t.Logger)
	farmRepo := persistence.NewFarmRepository(t.Logger)
	locationRepo := persistence.NewLocationRepository(t.Logger)
	productRepo := persistence.NewProductRepository(t.Logger)
	postRepo := persistence.NewPostRepository(t.Logger)
	transactionRepo := persistence.NewTransactionRepository(t.Logger)

	/* Service */

	authService := aAuth.NewService(t.Logger, transactor, generator, jwt, t.Redis, t.Queue, authRepo, userRepo)
	if t.Config.Auth.ExposeCodes {
		authService.ExposeCodes(true)
		t.Logger.Warn(context.Background(), shared.System, "EXPOSE_CODES is on: one-time codes are returned in API responses. Development only.")
	}
	userService := aUser.NewService(transactor, userRepo, authRepo, noteRepo)
	farmService := aFarm.NewService(transactor, userRepo, farmRepo)
	locationService := aLocation.NewService(transactor, userRepo, farmRepo, locationRepo)
	productService := aProduct.NewService(transactor, userRepo, farmRepo, productRepo, noteRepo)
	postService := aPost.NewService(transactor, userRepo, postRepo, noteRepo)
	transactionService := aTransaction.NewService(transactor, userRepo, farmRepo, productRepo, transactionRepo)

	return &Container{
		JWT:         jwt,
		Queue:       t.Queue,
		Auth:        &AuthModule{Service: authService},
		User:        &UserModule{Service: userService},
		Farm:        &FarmModule{Service: farmService},
		Location:    &LocationModule{Service: locationService},
		Product:     &ProductModule{Service: productService},
		Post:        &PostModule{Service: postService},
		Transaction: &TransactionModule{Service: transactionService},
	}
}

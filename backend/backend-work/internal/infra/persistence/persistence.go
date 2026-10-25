package persistence

import (
	"agropal/pkg"
)

// RepositoryImpl is the "Swiss Army Knife" that handles all SQL logic
type RepositoryImpl struct {
	logger *pkg.Logger
}

// NewRepository initializes the implementation
func NewRepository(logger *pkg.Logger) *RepositoryImpl {
	return &RepositoryImpl{logger: logger}
}

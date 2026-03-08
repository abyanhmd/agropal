package user

import dUser "agropal/internal/domain/user"

func ToUserModel(domain dUser.User) UserModel {
	return UserModel{
		ID: domain.ID,

		Name:     domain.Account.Name,
		Email:    domain.Account.Email,
		Username: domain.Account.Username,
		Avatar:   domain.Account.Avatar,

		Password: domain.Security.Password,
		Role:     domain.Security.Role,

		CreatedAt:   domain.Metadata.CreatedAt,
		UpdatedAt:   domain.Metadata.UpdatedAt,
		DeletedAt:   domain.Metadata.DeletedAt,
		LastLoginAt: domain.Metadata.LastLoginAt,
	}
}

func (m UserModel) ToDomain() dUser.User {
	return dUser.User{
		ID: m.ID,
		Account: dUser.AccountInfo{
			Name:     m.Name,
			Email:    m.Email,
			Username: m.Username,
			Avatar:   m.Avatar,
		},
		Security: dUser.SecurityInfo{
			Password: m.Password,
			Role:     m.Role,
		},
		Metadata: dUser.UserMetadata{
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			DeletedAt:   m.DeletedAt,
			LastLoginAt: m.LastLoginAt,
		},
	}
}

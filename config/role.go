package config

type Role string

const (
	RoleAdmin Role = "ADMIN" // System level
	RoleUser  Role = "USER"  // Default user
	RoleGuest Role = "GUEST" // Temporary/Viewer
)

// allPermissions is populated in init() from the role-scoped lists below so
// RoleAdmin doesn't need to duplicate every permission by hand.
var allPermissions []Permission

var userPermissions = []Permission{
	PermCreateFarm,
	PermGetOneFarm,
	PermGetManyFarms,
	PermUpdateFarm,
	PermDeleteFarm,

	PermCreatePost,
	PermGetOnePost,
	PermGetManyPosts,
	PermCreatePostComment,
	PermCreatePostEngagement,
	PermCreateCommentEngagement,

	PermCreateProduct,
	PermUpdateProduct,
	PermGetOneProduct,
	PermGetManyProducts,
	PermDeleteProduct,
	PermCreateProductStock,
	PermDeleteProductStock,
	PermCreateProductPrice,
	PermDeleteProductPrice,

	PermCreateTransaction,
	PermGetOneTransaction,
	PermGetManyTransactions,
	PermUpdateTransaction,
	PermDeleteTransaction,
	PermDeleteManyTransactions,
	PermDeleteAllTransactions,

	PermGetUserByID,
	PermUpdateUser,
	PermChangeUserPassword,
	PermDeleteUser,
	PermGetManyUsers,

	PermCreateWeatherLocation,
	PermGetWeatherLocationByFarmID,
	PermUpdateWeatherLocation,
	PermDeleteWeatherLocation,
}

var guestPermissions = []Permission{
	PermGetOneFarm,
	PermGetManyFarms,

	PermGetOnePost,
	PermGetManyPosts,

	PermGetOneProduct,
	PermGetManyProducts,

	PermGetWeatherLocationByFarmID,
}

// RolePermissions maps each role to the permissions it holds. RoleAdmin is
// filled in init() with the union of every known permission so it never
// falls out of sync when a new permission is added.
var RolePermissions = map[Role][]Permission{
	RoleUser:  userPermissions,
	RoleGuest: guestPermissions,
}

func init() {
	allPermissions = []Permission{
		PermCreateFarm, PermGetOneFarm, PermGetManyFarms, PermUpdateFarm, PermDeleteFarm,

		PermCreatePost, PermGetOnePost, PermGetManyPosts, PermCreatePostComment,
		PermCreatePostEngagement, PermCreateCommentEngagement,

		PermCreateProduct, PermUpdateProduct, PermGetOneProduct, PermGetManyProducts,
		PermDeleteProduct, PermCreateProductStock, PermDeleteProductStock,
		PermCreateProductPrice, PermDeleteProductPrice,

		PermCreateTransaction, PermGetOneTransaction, PermGetManyTransactions,
		PermUpdateTransaction, PermDeleteTransaction, PermDeleteManyTransactions,
		PermDeleteAllTransactions,

		PermGetUserByID, PermUpdateUser, PermChangeUserPassword, PermDeleteUser, PermGetManyUsers,

		PermCreateWeatherLocation, PermGetWeatherLocationByFarmID,
		PermUpdateWeatherLocation, PermDeleteWeatherLocation,
	}

	RolePermissions[RoleAdmin] = allPermissions
}

// HasPermission reports whether role has been granted perm.
func HasPermission(role Role, perm Permission) bool {
	for _, p := range RolePermissions[role] {
		if p == perm {
			return true
		}
	}
	return false
}

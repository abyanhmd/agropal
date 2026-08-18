package config

type Permission string

const (
	PermCreateFarm   Permission = "farm:create"
	PermGetOneFarm   Permission = "farm:get_one"
	PermGetManyFarms Permission = "farm:get_many"
	PermUpdateFarm   Permission = "farm:update"
	PermDeleteFarm   Permission = "farm:delete"

	PermCreatePost              Permission = "post:create"
	PermGetOnePost              Permission = "post:get_one"
	PermGetManyPosts            Permission = "post:get_many"
	PermCreatePostComment       Permission = "post:create_comment"
	PermCreatePostEngagement    Permission = "post:create_post_engagement"
	PermCreateCommentEngagement Permission = "post:create_email_engagement"

	PermCreateProduct      Permission = "product:create"
	PermUpdateProduct      Permission = "product:update"
	PermGetOneProduct      Permission = "product:get_one"
	PermGetManyProducts    Permission = "product:get_many"
	PermDeleteProduct      Permission = "product:delete"
	PermCreateProductStock Permission = "product:create_stock"
	PermDeleteProductStock Permission = "product:delete_stock"
	PermCreateProductPrice Permission = "product:create_price"
	PermDeleteProductPrice Permission = "product:delete_price"

	PermCreateTransaction      Permission = "transaction:create"
	PermGetOneTransaction      Permission = "transaction:get_one"
	PermGetManyTransactions    Permission = "transaction:get_many"
	PermUpdateTransaction      Permission = "transaction:update"
	PermDeleteTransaction      Permission = "transaction:delete"
	PermDeleteManyTransactions Permission = "transaction:delete_many"
	PermDeleteAllTransactions  Permission = "transaction:delete_all"

	PermGetUserByID        Permission = "user:get_one"
	PermUpdateUser         Permission = "user:update"
	PermChangeUserPassword Permission = "user:change_password"
	PermDeleteUser         Permission = "user:delete"
	PermGetManyUsers       Permission = "user:get_many"

	PermCreateWeatherLocation      Permission = "weather:create_location"
	PermGetWeatherLocationByFarmID Permission = "weather:get_location_by_farm_id"
	PermUpdateWeatherLocation      Permission = "weather:update_location"
	PermDeleteWeatherLocation      Permission = "weather:delete_location"
)

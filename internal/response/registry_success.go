package response

import (
	dAuth "agropal/internal/domain/auth"
	dFarm "agropal/internal/domain/farm"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
)

var SuccessRegistry = map[shared.SuccessKey]ResponseDetail{
	// COMMON
	shared.SuccessFileExported: {HTTPCode: 200, Message: "The file was exported successfully."}, // "COMMON.FILE_EXPORTED"

	// AUTH
	dAuth.SuccessRegister: {HTTPCode: 201, Message: "Registration successful."},      // "AUTH.REGISTER_SUCCESS"
	dAuth.SuccessLogin:    {HTTPCode: 200, Message: "Login successful."},             // "AUTH.LOGIN_SUCCESS"
	dAuth.SuccessLogout:   {HTTPCode: 200, Message: "Logout successful."},            // "AUTH.LOGOUT_SUCCESS"
	dAuth.SuccessRefresh:  {HTTPCode: 200, Message: "Token refreshed successfully."}, // "AUTH.REFRESH_SUCCESS"

	// FARM
	dFarm.SuccessCreated: {HTTPCode: 201, Message: "The farm was created successfully."}, // "FARM.CREATE_SUCCESS"
	dFarm.SuccessUpdated: {HTTPCode: 200, Message: "The farm was updated successfully."}, // "FARM.UPDATE_SUCCESS"
	dFarm.SuccessDeleted: {HTTPCode: 200, Message: "The farm was deleted successfully."}, // "FARM.DELETE_SUCCESS"

	// NOTE
	dNote.SuccessCreated: {HTTPCode: 201, Message: "The note was created successfully."}, // "NOTE.CREATE_SUCCESS"
	dNote.SuccessUpdated: {HTTPCode: 200, Message: "The note was updated successfully."}, // "NOTE.UPDATE_SUCCESS"
	dNote.SuccessDeleted: {HTTPCode: 200, Message: "The note was deleted successfully."}, // "NOTE.DELETE_SUCCESS"

	// POST
	dPost.SuccessCreated: {HTTPCode: 201, Message: "The post was created successfully."}, // "POST.CREATE_SUCCESS"
	dPost.SuccessUpdated: {HTTPCode: 200, Message: "The post was updated successfully."}, // "POST.UPDATE_SUCCESS"
	dPost.SuccessDeleted: {HTTPCode: 200, Message: "The post was deleted successfully."}, // "POST.DELETE_SUCCESS"

	dPost.SuccessCommentCreated: {HTTPCode: 201, Message: "The comment was created successfully."}, // "POST.COMMENT_CREATE_SUCCESS"
	dPost.SuccessCommentUpdated: {HTTPCode: 200, Message: "The comment was updated successfully."}, // "POST.COMMENT_UPDATE_SUCCESS"
	dPost.SuccessCommentDeleted: {HTTPCode: 200, Message: "The comment was deleted successfully."}, // "POST.COMMENT_DELETE_SUCCESS"

	// PRODUCT
	dProduct.SuccessCreated: {HTTPCode: 201, Message: "The product was created successfully."}, // "PRODUCT.CREATE_SUCCESS"
	dProduct.SuccessUpdated: {HTTPCode: 200, Message: "The product was updated successfully."}, // "PRODUCT.UPDATE_SUCCESS"
	dProduct.SuccessDeleted: {HTTPCode: 200, Message: "The product was deleted successfully."}, // "PRODUCT.DELETE_SUCCESS"

	dProduct.SuccessStockAdded:   {HTTPCode: 201, Message: "The stock movement was recorded successfully."}, // "PRODUCT.STOCK_ADD_SUCCESS"
	dProduct.SuccessStockDeleted: {HTTPCode: 200, Message: "The stock movement was deleted successfully."},  // "PRODUCT.STOCK_DELETE_SUCCESS"

	dProduct.SuccessPriceAdded:   {HTTPCode: 201, Message: "The price was added successfully."},   // "PRODUCT.PRICE_ADD_SUCCESS"
	dProduct.SuccessPriceDeleted: {HTTPCode: 200, Message: "The price was deleted successfully."}, // "PRODUCT.PRICE_DELETE_SUCCESS"

	// TRANSACTION
	dTransaction.SuccessCreated: {HTTPCode: 201, Message: "The transaction was created successfully."}, // "TRANSACTION.CREATE_SUCCESS"
	dTransaction.SuccessUpdated: {HTTPCode: 200, Message: "The transaction was updated successfully."}, // "TRANSACTION.UPDATE_SUCCESS"
	dTransaction.SuccessDeleted: {HTTPCode: 200, Message: "The transaction was deleted successfully."}, // "TRANSACTION.DELETE_SUCCESS"

	// USER
	dUser.SuccessUpdated: {HTTPCode: 200, Message: "The user was updated successfully."}, // "USER.UPDATE_SUCCESS"
	dUser.SuccessDeleted: {HTTPCode: 200, Message: "The user was deleted successfully."}, // "USER.DELETE_SUCCESS"

	// WEATHER
	dWeather.SuccessCreated: {HTTPCode: 201, Message: "The weather data was created successfully."}, // "WEATHER.CREATE_SUCCESS"
	dWeather.SuccessUpdated: {HTTPCode: 200, Message: "The weather data was updated successfully."}, // "WEATHER.UPDATE_SUCCESS"
	dWeather.SuccessDeleted: {HTTPCode: 200, Message: "The weather data was deleted successfully."}, // "WEATHER.DELETE_SUCCESS"
}

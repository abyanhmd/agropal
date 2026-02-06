package shared

type AssetType string
type DeleteType string

const (
	System           AssetType = "System"
	AssetUser        AssetType = "User"
	AssetFarm        AssetType = "Farm"
	AssetProduct     AssetType = "Product"
	AssetTransaction AssetType = "Transaction"
	AssetWeather     AssetType = "Weather"
	AssetPost        AssetType = "Post"
	AssetComment     AssetType = "Comment"
)

const (
	DeleteSoft DeleteType = "soft"
	DeleteHard DeleteType = "hard"
)

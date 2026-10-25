package location

import (
	dLocation "agropal/internal/domain/location"
	"agropal/internal/response"
)

func DomainToLocationResponse(location dLocation.Location) response.LocationResponse {
	return response.LocationResponse{
		LocationID: location.LocationID,
		Longitude:  location.Longitude,
		Latitude:   location.Latitude,
		FarmID:     location.FarmID,
	}
}

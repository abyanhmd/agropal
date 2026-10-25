package request

import (
	aLocation "agropal/internal/app/location"
)

type LocationRequest struct {
	Longitude float64 `json:"longitude" validate:"longitude"`
	Latitude  float64 `json:"latitude" validate:"latitude"`
}

func (r *LocationRequest) ToCommand() aLocation.LocationCommand {
	return aLocation.LocationCommand{
		Longitude: r.Longitude,
		Latitude:  r.Latitude,
	}
}

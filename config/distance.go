package config

import (
	"fmt"
	"math"
)

const radioTierraKm = 6378.0

func degreesToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {

	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	differenceLon := lon2Rad - lon1Rad
	differenceLat := lat2Rad - lat1Rad

	a := math.Pow(math.Sin(differenceLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(differenceLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return radioTierraKm * c
}

func radioAllow(lat2, lon2 float64) bool {
	lat1 := -17.783309
	lon1 := -63.182122
	distance := calculateDistance(lat1, lon1, lat2, lon2)
	fmt.Println("La distancia es ", distance)
	return distance >= 519.93
}

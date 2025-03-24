package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/umahmood/haversine"
	"net/http"
	"os"
)

func getIPLocation(ip string) (haversine.Coord, error) {
	ACCESS_KEY_IP_API := os.Getenv("ACCESS_KEY_IP_API")
	resp, err := http.Get(
		fmt.Sprintf(
			"https://api.ipapi.com/api/%v?access_key=%v",
			ip, ACCESS_KEY_IP_API,
		),
	)
	if err != nil {
		return haversine.Coord{}, err
	}
	defer resp.Body.Close()

	var data struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return haversine.Coord{}, err
	}

	return haversine.Coord{Lat: data.Latitude, Lon: data.Longitude}, nil
}

// @Summary      Calculate Distance
// @Description  Calculate Distance from your location to a city
// @Param        city   query  string     false   "City"
// @Success      200  {array}  map[string]interface{}
// @Router       /distance [get]
func calculateDistance(c *gin.Context) {
	ip := c.ClientIP()
	city := c.Query("city")

	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Missing city",
		})
		return
	}

	userLocation, err := getIPLocation(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get IP location"})
		return
	}

	cityLocation, exists := cityCoordinates[city]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City not found no data yet"})
		return
	}

	distance, _ := haversine.Distance(userLocation, cityLocation)

	c.JSON(http.StatusOK, gin.H{
		"ip":       ip,
		"user_lat": userLocation.Lat,
		"user_lon": userLocation.Lon,
		"city":     city,
		"city_lat": cityLocation.Lat,
		"city_lon": cityLocation.Lon,
		"distance": fmt.Sprintf("%.2f km", distance),
	})
}

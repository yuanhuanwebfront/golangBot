package services

import (
	"encoding/json"
	"fmt"
	"github.com/luckfunc/golangBot/internal/models"
	"net/http"
)

const weatherAPIKey = "62a43a5bf3944068978fd939241d2dba"

func GetWeather(locationId string) (models.WeatherNow, error) {
	baseUrl := "https://devapi.qweather.com/v7/weather/now"
	urlWithParams := fmt.Sprintf("%s?location=%s&key=%s", baseUrl, locationId, weatherAPIKey)

	resp, err := http.Get(urlWithParams)
	if err != nil {
		return models.WeatherNow{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.WeatherNow{}, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	var weatherInfo models.WeatherInfo
	if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
		return models.WeatherNow{}, err
	}

	return weatherInfo.Now, nil
}

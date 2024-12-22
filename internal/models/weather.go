package models

type WeatherNow struct {
	ObsTime string `json:"obsTime"`
	Temp    string `json:"temp"`
	Text    string `json:"text"`
}

type WeatherInfo struct {
	Code       string     `json:"code"`
	UpdateTime string     `json:"updateTime"`
	Now        WeatherNow `json:"now"`
}

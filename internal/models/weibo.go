package models

type WeiboHotList struct {
	Success bool         `json:"success"`
	Time    string       `json:"time"`
	Data    []WeiboEntry `json:"data"`
}

type WeiboEntry struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Hot   string `json:"hot"`
}

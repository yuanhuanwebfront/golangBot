package services

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func HandleDouYinLink(msg *openwechat.Message) {
	videoUrl, err := getDouYinVideoUrl(msg.Content)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := downloadDouYinVideo(videoUrl, msg); err != nil {
		fmt.Println("Error:", err)
	}
}

func getDouYinVideoUrl(videoUrl string) (string, error) {
	cleanedVideoUrl := url.QueryEscape(videoUrl)
	baseUrl := "http://localhost/api/download?url=" + cleanedVideoUrl + "&prefix=False&watermark=False"

	resp, err := http.Get(baseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

func downloadDouYinVideo(videoUrl string, msg *openwechat.Message) error {
	videoResponse, err := http.Get(videoUrl)
	if err != nil {
		return err
	}
	defer videoResponse.Body.Close()

	fileName := "./assets/douyin/douyin" + strconv.FormatInt(time.Now().Unix(), 10) + ".mp4"
	videoFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer videoFile.Close()

	if _, err = io.Copy(videoFile, videoResponse.Body); err != nil {
		return err
	}

	video, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer video.Close()

	msg.ReplyVideo(video)
	return nil
}

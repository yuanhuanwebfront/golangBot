package main

import (
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // Desktop mode

	// Register QR code callback
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// Create hot reload storage object
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()

	// Perform hot login
	if err := performHotLogin(bot, reloadStorage); err != nil {
		fmt.Println(err)
		return
	}

	// Handle messages
	bot.MessageHandler = handleMessage

	// Block the main goroutine until an exception occurs or the user exits
	bot.Block()
}

func performHotLogin(bot *openwechat.Bot, reloadStorage openwechat.HotReloadStorage) error {
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		return err
	}
	self, err := bot.GetCurrentUser()
	if err != nil {
		return err
	}
	groups, err := self.Groups()
	fmt.Println(groups, err)
	return nil
}

func handleMessage(msg *openwechat.Message) {
	sender, _ := msg.SenderInGroup()

	fmt.Println(sender, msg.Content)

	if msg.IsText() && msg.Content == "ping" {
		msg.ReplyText("pong")
	}

	if msg.IsText() && msg.Content == "摸鱼" {
		if err := downloadMoYuImageAndReply(msg); err != nil {
			fmt.Println("Error:", err)
		}
	}
	// 检查消息是否来自群聊
	if msg.IsSendByGroup() {
		// 如果是群聊消息，检查消息内容是否为 "@铲车司机bot 摸鱼日历"
		if msg.Content == "@铲车司机bot 摸鱼日历" {
			// 如果消息内容匹配，触发处理函数
			HandleGroupChatMessage(msg)
		}

		if msg.Content == "随机" {
			handleGroupChatImage(msg)
		}

		if msg.Content == "轩子" {
			handleGroupChatXuanZiImgae(msg)
		}

	}
}

func downloadMoYuImageAndReply(msg *openwechat.Message) error {

	imgUrl, err := getMoyuImageURL()
	if err != nil {
		return err
	}
	imgResponse, err := http.Get(imgUrl)
	if err != nil {
		return err
	}
	defer imgResponse.Body.Close()

	// 创建本地文件保存
	imgFile, err := os.Create("./image/moyu.png")

	if err != nil {
		return err
	}

	defer imgFile.Close()

	// 将图像内容保存到本地文件
	_, err = io.Copy(imgFile, imgResponse.Body)
	if err != nil {
		fmt.Println("Error while saving the image:", err)
		return nil
	}

	img, err := os.Open("./image/moyu.png")
	if err != nil {
		return err
	}
	defer img.Close()

	msg.ReplyImage(img)
	return nil
}

// 获取摸鱼图片的URL
func getMoyuImageURL() (string, error) {
	resp, err := http.Get("https://api.vvhan.com/api/moyu?type=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var moyuResp struct {
		Success bool   `json:"success"`
		URL     string `json:"url"`
	}

	err = json.Unmarshal(body, &moyuResp)
	if err != nil {
		return "", err
	}

	return moyuResp.URL, nil
}

// HandleGroupChatMessage 处理群聊消息
func HandleGroupChatMessage(msg *openwechat.Message) {
	// 在这里编写处理群聊消息的代码
	if err := downloadMoYuImageAndReply(msg); err != nil {
		fmt.Println("Error:", err)
	}
}

// 发送图片
func handleGroupChatImage(msg *openwechat.Message) {
	// 指定本地图片文件夹的路径
	imageDir := "totalImage" // "image" 是您的图片文件夹的相对路径

	// 读取图片文件列表
	imageFiles, err := ioutil.ReadDir(imageDir)
	if err != nil {
		fmt.Println("Error reading image directory:", err)
		return
	}

	// 创建一个存储图片文件路径的切片
	imagePaths := []string{}

	// 遍历图片文件列表，将图片文件的路径添加到切片中
	for _, fileInfo := range imageFiles {
		if !fileInfo.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(imageDir, fileInfo.Name()))
		}
	}

	// 随机选择一张图片
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(imagePaths))
	selectedImagePath := imagePaths[randomIndex]

	fmt.Println("selectedImagePath", selectedImagePath)
	file, err := os.Open(selectedImagePath)
	if err != nil {
		fmt.Println("Error opening image file:", err)
		return
	}
	defer file.Close()

	_, err = msg.ReplyImage(file)
	if err != nil {
		fmt.Println("Error sending image:", err)
		return
	}

	//fmt.Println("Sending image:", selectedImagePath)

}

// 发送轩子
func handleGroupChatXuanZiImgae(msg *openwechat.Message) {
	// 指定本地图片文件夹的路径
	imageDir := "轩子巨2兔" // "image" 是您的图片文件夹的相对路径

	// 读取图片文件列表
	imageFiles, err := ioutil.ReadDir(imageDir)
	if err != nil {
		fmt.Println("Error reading image directory:", err)
		return
	}

	// 创建一个存储图片文件路径的切片
	imagePaths := []string{}

	// 遍历图片文件列表，将图片文件的路径添加到切片中
	for _, fileInfo := range imageFiles {
		if !fileInfo.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(imageDir, fileInfo.Name()))
		}
	}

	// 随机选择一张图片
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(imagePaths))
	selectedImagePath := imagePaths[randomIndex]

	fmt.Println("selectedImagePath", selectedImagePath)
	file, err := os.Open(selectedImagePath)
	if err != nil {
		fmt.Println("Error opening image file:", err)
		return
	}
	defer file.Close()

	_, err = msg.ReplyImage(file)
	if err != nil {
		fmt.Println("Error sending image:", err)
		return
	}

	//fmt.Println("Sending image:", selectedImagePath)

}

func handleGroupChatTenImage(msg *openwechat.Message, numberOfImages int) {
	// 指定本地图片文件夹的路径
	imageDir := "totalImage" // "image" 是您的图片文件夹的相对路径

	// 读取图片文件列表
	imageFiles, err := ioutil.ReadDir(imageDir)
	if err != nil {
		fmt.Println("Error reading image directory:", err)
		return
	}

	// 创建一个存储图片文件路径的切片
	imagePaths := []string{}

	// 遍历图片文件列表，将图片文件的路径添加到切片中
	for _, fileInfo := range imageFiles {
		if !fileInfo.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(imageDir, fileInfo.Name()))
		}
	}

	// 随机选择并发送指定数量的图片
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numberOfImages; i++ {
		randomIndex := rand.Intn(len(imagePaths))
		selectedImagePath := imagePaths[randomIndex]

		fmt.Println("selectedImagePath", selectedImagePath)
		file, err := os.Open(selectedImagePath)
		if err != nil {
			fmt.Println("Error opening image file:", err)
			return
		}
		defer file.Close()

		_, err = msg.ReplyImage(file)
		if err != nil {
			fmt.Println("Error sending image:", err)
			return
		}
	}
}

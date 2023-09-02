package main

import (
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type MyStruct struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		MoyuURL string `json:"moyu_url"`
	} `json:"data"`
}

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()

	// 执行热登录
	if err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption()); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
		fmt.Println("text", msg.Content)

		if msg.IsTickledMe() {
			replyText := "我是一个机器人"
			msg.ReplyText(replyText)
		}

		if msg.IsText() && msg.Content == "摸鱼" {
			response, err := http.Get("https://api.j4u.ink/v1/store/other/proxy/remote/moyu.json")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			defer response.Body.Close() // 在函数退出时关闭响应体

			if response.StatusCode != http.StatusOK {
				fmt.Println("Error response from server:", response.Status)
				return
			}

			// 读取响应体数据
			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}

			// 解析 JSON 数据
			var result MyStruct
			if err := json.Unmarshal(data, &result); err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}
			// 访问解析后的字段
			imgUrl := result.Data.MoyuURL

			// 下载图像到本地
			imgResponse, err := http.Get(imgUrl)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			defer imgResponse.Body.Close()

			// 创建本地文件来保存图像
			imgFile, err := os.Create("./image/moyu.png")
			if err != nil {
				fmt.Println("Error while creating the file:", err)
				return
			}
			defer imgFile.Close()
			// 将图像内容保存到本地文件
			_, err = io.Copy(imgFile, imgResponse.Body)
			if err != nil {
				fmt.Println("Error while saving the image:", err)
				return
			}

			// 发送本地图像文件
			img, _ := os.Open("./image/moyu.png")
			defer img.Close()
			msg.ReplyImage(img)
		}
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

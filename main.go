package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	msgId := msg.MsgId
	msg.Set(msgId, msg.Content)

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
		fmt.Println(msg.MsgId, msg.Content)

		if msg.IsTickledMe() || msg.Content == "@铲车司机bot " {
			msg.ReplyText("您好！我是铲车司机Bot，我可以为您提供以下功能：\n" +
				"1. 回复 'ping' 获取 'pong' 响应。\n" +
				"2. 回复 '摸鱼' 获取随机摸鱼图片。\n" +
				"3. 回复 '@铲车司机bot 摸鱼日历' 获取摸鱼日历。\n" +
				"4. 回复 '随机' 获取随机图片。\n" +
				"5. 回复 '轩子' 获取轩子图片。\n" +
				"6. 回复 '微博热搜' 或 '热搜' 获取微博热搜内容。\n" +
				"7. 回复 '简报' 获取最新简报图片。\n" +
				// 在这里添加其他功能的介绍
				// 例如："8. 回复 '其他功能关键词' 获取其他功能介绍。\n"+
				"请随时使用这些关键词与我互动，我将尽力为您提供帮助！")
		}

		if strings.Contains(msg.Content, "信念力") {
			msg.ReplyText("信念力\n" +
				"你知道是啥吗\n" +
				"能一起把生活过好\n" +
				"不是你想的都是钱堆出来的普通人哪有那么多钱")
		}

		//如果消息是以@铲车司机bot  开头的
		// 如果消息是包含douyin.com的链接
		if strings.Contains(msg.Content, "douyin.com") {
			fmt.Println("msg", msg)
			// 去除前缀留下后面的内容
			msgVideoUrl := strings.TrimPrefix(msg.Content, "@铲车司机bot ")

			fmt.Println("msgVideoUrl", msgVideoUrl)

			// 获取videoUrl
			video, err := getDouYinVideoUrl(msgVideoUrl)
			fmt.Println("video", video)

			if err != nil {
				fmt.Println("Error:", err)
			}
			// 下载视频到本地download文件
			if err := downloadDouYinVideo(video, msg); err != nil {
				fmt.Println("Error:", err)
			}
		}
		if strings.Contains(msg.Content, "劝人") {
			msg.ReplyText("劝人就5分钟。5分钟没说动的事，就不再劝了。而是应该想想，捆住他手脚的是什么。是什么把他压在那里，让他没办法往前走。\n所以，我看到有人在一个没啥前途的公司岗位上待着不辞职不转行，在一段没有爱的关系里呆着被折磨，而不离开，不是对方对她好，有承诺，而是对方抓住了她的恐惧。\n所以，知道道理，依然过不好这一生。")
		}
		// 如果是群聊消息，检查消息内容是否为 "@铲车司机bot 摸鱼日历"
		if msg.Content == "@铲车司机bot 摸鱼日历" {
			// 如果消息内容匹配，触发处理函数
			HandleGroupChatMessage(msg)
		}

		if msg.Content == "随机" || strings.Contains(msg.Content, "色图") {
			handleGroupChatImage(msg)
		}

		if msg.Content == "轩子" {
			handleGroupChatXuanZiImgae(msg)
		}

		if msg.Content == "微博热搜" || msg.Content == "热搜" {
			handleGetWeiBoHotList(msg)
		}

		if msg.Content == "简报" {
			downloadNewsImageAndReply(msg)
		}

		if strings.Contains(msg.Content, "电动车") {
			msg.ReplyText("妈的, 在西安花4000买个电动车")
		}

		if strings.Contains(msg.Content, "天气") {
			// 请输入城市和区域信息 例如：回复 "西安 雁塔" 获取西安雁塔区的天气信息
			msg.ReplyText("请输入城市和区域信息例如: " + "\n" +
				"回复 \"!西安 雁塔\" 获取西安雁塔区的实时天气信息")
		}

		if strings.HasPrefix(msg.Content, "!") {
			parts := strings.Split(msg.Content, " ")
			// 去除感叹号前缀
			if len(parts) > 0 && strings.HasPrefix(parts[0], "!") {
				parts[0] = strings.TrimPrefix(parts[0], "!")
			}
			if len(parts) < 2 {
				cityList, _ := getCityList(parts[0])
				replyMessage := "输入有误,请输入! " + parts[0] + "然后加上以下城市信息例如：\n"
				for i, cityInfo := range cityList {
					replyMessage += "!" + parts[0] + " " + cityInfo.Name
					// 如果不是最后一个城市信息，则添加换行符
					if i < len(cityList)-1 {
						replyMessage += "\n"
					}
				}
				msg.ReplyText(replyMessage)
				return
			}
			// 传入城市信息和区域信息
			locationId, err := getCityInfo(parts[0], parts[1])
			if err != nil {
				return
			}

			if err != nil {
				msg.ReplyText("请输入正确的城市和区域信息")
				return
			}
			weatherInfo, _ := getWeather(locationId)
			// 长安区的当天天气信息
			msg.ReplyText(parts[1] + "的天气: " + "\n" +
				"实时天气: " + weatherInfo.Text + "\n" +
				"实时温度: " + weatherInfo.Temp + "℃")
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

// 获取微博热搜
func handleGetWeiBoHotList(msg *openwechat.Message) (*WeiboHotList, error) {
	resp, err := http.Get("https://api.vvhan.com/api/wbhot")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解码JSON数据到 WeiboHotList 结构体
	var weiboHotList WeiboHotList
	err = json.Unmarshal(body, &weiboHotList)
	if err != nil {
		return nil, err
	}
	result := weiboHotList
	fmt.Println("result", result.Data)

	// 构建最终的文本
	var resultText strings.Builder
	resultText.WriteString("微博热搜: " + weiboHotList.Time + "\n")
	// 遍历前十条数据并添加到文本中

	for i, entry := range weiboHotList.Data {
		hotValue, _ := strconv.Atoi(entry.Hot)
		if i >= 5 {
			break // 仅遍历前十条数据
		}
		resultText.WriteString(fmt.Sprintf("%d. 标题: %s\n   链接: %s\n   热度: %d 万\n\n", i+1, entry.Title, entry.URL, hotValue/10000))
	}
	// 输出整个文本
	fmt.Println(resultText.String())
	msg.ReplyText(resultText.String())

	// 返回解码后的结构体
	return &weiboHotList, nil
}

// 获取新闻接口
func getNewNewsUrls() (string, error) {
	resp, err := http.Get("https://dayu.qqsuu.cn/weiyujianbao/apis.php?type=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var newNewsResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Url  string `json:"data"`
	}

	err = json.Unmarshal(body, &newNewsResp)
	if err != nil {
		fmt.Println("err", err)
		return "", err
	}
	// 将 data 字段的值作为函数返回值
	return newNewsResp.Url, nil
}

// 下载新闻图片
func downloadNewsImageAndReply(msg *openwechat.Message) error {

	imgUrl, err := getNewNewsUrls()
	if err != nil {
		return err
	}
	imgResponse, err := http.Get(imgUrl)
	if imgResponse.StatusCode != http.StatusOK {
		message := fmt.Sprintf("HTTP请求失败，状态码：%d", imgResponse.StatusCode)
		msg.ReplyText(message)
		return errors.New(message)
	}

	if err != nil {
		return err
	}
	defer imgResponse.Body.Close()

	// 创建本地文件保存
	imgFile, err := os.Create("./image/news.png")

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

	img, err := os.Open("./image/news.png")
	if err != nil {
		return err
	}
	defer img.Close()

	msg.ReplyImage(img)
	return nil
}

// 获取实时天气接口 https://devapi.qweather.com.qweather.com/v7/weather/now?
// key=62a43a5bf3944068978fd939241d2dba
// location必填
type WeatherNow struct {
	ObsTime string `json:"obsTime"`
	Temp    string `json:"temp"` // temp 实时温度
	Text    string `json:"text"` // 天气状况的文字描述，包括阴晴雨雪等天气状态的描述
}
type WeatherInfo struct {
	Code       string     `json:"code"`
	UpdateTime string     `json:"updateTime"`
	Now        WeatherNow `json:"now"`
}

func getWeather(locationId string) (WeatherNow, error) {
	baseUrl := "https://devapi.qweather.com/v7/weather/now"
	urlWithParams := baseUrl + "?location=" + locationId + "&key=62a43a5bf3944068978fd939241d2dba"

	resp, err := http.Get(urlWithParams)
	if err != nil {
		return WeatherNow{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherNow{}, fmt.Errorf("HTTP请求失败，状态码：%d", resp.StatusCode)
	}
	var weatherInfo WeatherInfo
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&weatherInfo); err != nil {
		return WeatherNow{}, err
	}

	return weatherInfo.Now, err
}

type CityInfo struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
	Adm2      string `json:"adm2"`
	Adm1      string `json:"adm1"`
	Country   string `json:"country"`
	TimeZone  string `json:"tz"`
	Type      string `json:"type"`
}

type CityInfoResponse struct {
	Code     string     `json:"code"`
	Location []CityInfo `json:"location"`
}

// 获取城市信息通过location, 例如西安, url = https://geoapi.qweather.com/v2/city/lookup?[请求参数]
func getCityInfo(city string, area string) (string, error) {
	baseUrl := "https://geoapi.qweather.com/v2/city/lookup"
	encodedLocation := url.QueryEscape(city)
	urlWithParams := baseUrl + "?location=" + encodedLocation + "&key=62a43a5bf3944068978fd939241d2dba"

	resp, err := http.Get(urlWithParams)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码：%d", resp.StatusCode)
	}

	var cityInfoResponse CityInfoResponse
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&cityInfoResponse); err != nil {
		return "", err
	}

	// 遍历城市信息，将与area匹配的城市的ID存储在selectedCityIDs切片中
	for _, cityInfo := range cityInfoResponse.Location {
		if cityInfo.Name == area {
			return cityInfo.Id, nil
		}
	}

	return "", nil
}

// 获取城市列表
func getCityList(city string) ([]CityInfo, error) {
	baseUrl := "https://geoapi.qweather.com/v2/city/lookup"
	encodedLocation := url.QueryEscape(city)
	fmt.Println("citty", city)

	urlWithParams := baseUrl + "?location=" + encodedLocation + "&key=62a43a5bf3944068978fd939241d2dba"

	resp, err := http.Get(urlWithParams)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败，状态码：%d", resp.StatusCode)
	}

	var cityInfoResponse CityInfoResponse
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&cityInfoResponse); err != nil {
		return nil, err
	}
	// 去除重复的城市信息
	uniqueCities := make(map[string]CityInfo)
	for _, cityInfo := range cityInfoResponse.Location {
		uniqueCities[cityInfo.Name] = cityInfo
	}

	uniqueCityList := make([]CityInfo, 0, len(uniqueCities))
	for _, cityInfo := range uniqueCities {
		uniqueCityList = append(uniqueCityList, cityInfo)
	}

	return uniqueCityList, nil
}

// 获取抖音视频无水印链接
// 从本地localhost:8000/download
func getDouYinVideoUrl(videoUrl string) (string, error) {
	cleanedVideoUrl := url.QueryEscape(videoUrl)

	baseUrl := "http://localhost:8000/download?url=" + cleanedVideoUrl + "&prefix=False&watermark=False"
	fmt.Println("baseUrl", baseUrl)

	resp, err := http.Get(baseUrl)
	fmt.Println("http://localhost:8000/download?url="+videoUrl+"&prefix=False&watermark=False", "?????")

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

// 下载抖音视频
func downloadDouYinVideo(videoUrl string, msg *openwechat.Message) error {
	videoResponse, err := http.Get(videoUrl)
	if err != nil {
		return err
	}
	defer videoResponse.Body.Close()

	// 创建本地文件保存, 文件名称douyin + 时间戳
	fileName := "./download/douyin" + strconv.FormatInt(time.Now().Unix(), 10) + ".mp4"
	videoFile, err := os.Create(fileName)

	if err != nil {
		return err
	}
	defer videoFile.Close()

	//将视频内容保存到本地文件
	_, err = io.Copy(videoFile, videoResponse.Body)

	if err != nil {
		fmt.Println("Error while saving the video:", err)
		return nil
	}

	video, err := os.Open(fileName)

	if err != nil {
		fmt.Println("err", err)

		return err
	}
	defer video.Close()

	msg.ReplyVideo(video)
	return nil
}

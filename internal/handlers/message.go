package handlers

import (
	"github.com/eatmoreapple/openwechat"
	"github.com/luckfunc/golangBot/internal/services"
	"strings"
)

func HandleGroupMessage(msg *openwechat.Message) {
	if !msg.IsSendByGroup() {
		return
	}

	// 处理特殊关键字和回复
	handleSpecialKeywords(msg)

	// 处理抖音链接
	if strings.Contains(msg.Content, "douyin.com") {
		services.HandleDouYinLink(msg)
	}
}

func handleSpecialKeywords(msg *openwechat.Message) {
	switch {
	case msg.IsTickledMe() || msg.Content == "@铲车司机bot ":
		sendHelpMessage(msg)
	case strings.Contains(msg.Content, "信念力"):
		msg.ReplyText("信念力\n你知道是啥吗\n能一起把生活过好\n不是你想的都是钱堆出来的普通人哪有那么多钱")
	case strings.Contains(msg.Content, "劝人"):
		msg.ReplyText("劝人就5分钟。5分钟没说动的事，就不再劝了...")
	case strings.Contains(msg.Content, "热搜"):
		services.HandleWeiboHotList(msg)
	case strings.Contains(msg.Content, "摸鱼"):
		services.HandleMoYuImage(msg)
		// ... 其他情况
	}
}

func sendHelpMessage(msg *openwechat.Message) {
	msg.ReplyText("您好！我是铲车司机Bot，我可以为您提供以下功能：\n" +
		"1. 回复 'ping' 获取 'pong' 响应。\n" +
		"2. 回复 '摸鱼' 获取随机摸鱼图片。\n" +
		"3. 回复 '@铲车司机bot 摸鱼日历' 获取摸鱼日历。\n" +
		"4. 回复 '微博热搜' 或 '热搜' 获取微博热搜内容。\n",
	)
}

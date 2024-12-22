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

	// 处理抖音链接
	if strings.Contains(msg.Content, "douyin.com") {
		services.HandleDouYinLink(msg)
	}
	// 如果发送的消息包含过年，则发送当前距离除夕还有多少天
	if strings.Contains(msg.Content, "过年") {
		services.HandleNewYearCountdown(msg)
	}
}

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

	// 处理股票查询，格式：股票600519
	if strings.HasPrefix(msg.Content, "股票") {
		// 去掉"股票"二字，保留后面的代码部分
		code := strings.TrimPrefix(msg.Content, "股票")
		code = strings.TrimSpace(code) // 去掉可能的空格
		if code != "" {
			msg.Content = code // 设置消息内容为纯代码
			services.HandleStockQuery(msg)
		}
	}

	// 处理大盘查询，当输入牛来了，或者牛跑了，或者牛回速归，则发送大盘概览
	if strings.Contains(msg.Content, "牛来了") || strings.Contains(msg.Content, "牛跑了") || strings.Contains(msg.Content, "牛回速归") {
		services.HandleMarketOverview(msg)
	}
}

func handleSpecialKeywords(msg *openwechat.Message) {
	switch {
	// ... 其他 case
	case strings.Contains(msg.Content, "股票"):
		services.HandleStockQuery(msg)
	}
}

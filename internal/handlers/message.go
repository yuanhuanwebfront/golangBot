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

	if msg.IsAt() && shouldReplyStockHelp(msg.Content) {
		services.HandleStockHelp(msg)
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

	// 处理股票相关指令
	if strings.HasPrefix(msg.Content, "股票") {
		services.HandleStockCommand(msg)
	}

	// 处理大盘查询，当输入牛来了，或者牛跑了，或者牛回速归，牛死速跑，则发送大盘概览
	if strings.Contains(msg.Content, "牛来了") || strings.Contains(msg.Content, "牛跑了") || strings.Contains(msg.Content, "牛回速归") || strings.Contains(msg.Content, "牛死速跑") {
		services.HandleMarketOverview(msg)
	}
}

func shouldReplyStockHelp(content string) bool {
	lower := strings.ToLower(content)
	if strings.Contains(content, "功能") || strings.Contains(content, "帮助") || strings.Contains(content, "指令") || strings.Contains(lower, "help") {
		return true
	}
	cleaned := strings.ReplaceAll(content, "\u2005", " ")
	fields := strings.Fields(cleaned)
	var remaining []string
	for _, field := range fields {
		if strings.HasPrefix(field, "@") {
			continue
		}
		remaining = append(remaining, field)
	}
	if len(remaining) == 0 {
		return true
	}
	return len(remaining) == 1 && remaining[0] == "股票"
}

func handleSpecialKeywords(msg *openwechat.Message) {
	switch {
	// ... 其他 case
	case strings.Contains(msg.Content, "股票"):
		services.HandleStockQuery(msg)
	}
}

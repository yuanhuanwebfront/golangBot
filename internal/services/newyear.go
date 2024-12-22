package services

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"time"
)

func HandleNewYearCountdown(msg *openwechat.Message) {
	// 计算出来当前距离除夕(2025 年 1 月 28 日)还有多少天
	chineseNewYear := time.Date(2025, time.January, 28, 0, 0, 0, 0, time.Local)
	days := chineseNewYear.Sub(time.Now()).Hours() / 24
	msg.ReplyText(fmt.Sprintf("当前距离除夕还有%d天", int(days)))
}

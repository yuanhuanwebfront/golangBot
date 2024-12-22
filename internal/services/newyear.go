package services

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"time"
)

// 定义春节信息结构
type SpringFestival struct {
	EveDate     time.Time // 除夕
	NewYearDate time.Time // 大年初一
	Year        string    // 农历年份
}

func HandleNewYearCountdown(msg *openwechat.Message) {
	// 定义未来几年的除夕和春节时间
	festivals := []SpringFestival{
		{
			time.Date(2024, time.February, 9, 0, 0, 0, 0, time.Local),
			time.Date(2024, time.February, 10, 0, 0, 0, 0, time.Local),
			"甲辰年（龙年）",
		},
		{
			time.Date(2025, time.January, 28, 0, 0, 0, 0, time.Local),
			time.Date(2025, time.January, 29, 0, 0, 0, 0, time.Local),
			"乙巳年（蛇年）",
		},
		{
			time.Date(2026, time.February, 16, 0, 0, 0, 0, time.Local),
			time.Date(2026, time.February, 17, 0, 0, 0, 0, time.Local),
			"丙午年（马年）",
		},
	}

	now := time.Now()
	var nextFestival SpringFestival

	// 找到最近的一个春节
	for _, festival := range festivals {
		if festival.EveDate.After(now) {
			nextFestival = festival
			break
		}
	}

	// 计算除夕和春节的天数差
	daysToEve := int(nextFestival.EveDate.Sub(now).Hours()/24) + 1
	daysToNewYear := int(nextFestival.NewYearDate.Sub(now).Hours()/24) + 1

	// 格式化消息
	message := fmt.Sprintf("🧧 农历新年倒计时 🧧\n\n"+
		"距离除夕还有 %d 天\n"+
		"除夕时间：%d年%d月%d日\n\n"+
		"距离大年初一还有 %d 天\n"+
		"春节时间：%d年%d月%d日\n\n"+
		"农历%s新年\n"+
		"愿新年群友们身体健康，万事如意！🎊",
		daysToEve,
		nextFestival.EveDate.Year(),
		nextFestival.EveDate.Month(),
		nextFestival.EveDate.Day(),
		daysToNewYear,
		nextFestival.NewYearDate.Year(),
		nextFestival.NewYearDate.Month(),
		nextFestival.NewYearDate.Day(),
		nextFestival.Year)

	msg.ReplyText(message)
}

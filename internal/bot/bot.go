package bot

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/luckfunc/golangBot/internal/handlers"
)

func Run() error {
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// Register QR code callback
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 直接登录，不使用热登录
	if err := bot.Login(); err != nil {
		return fmt.Errorf("login failed: %v", err)
	}

	// Handle group messages
	bot.MessageHandler = handlers.HandleGroupMessage

	// Block until exit
	bot.Block()
	return nil
}

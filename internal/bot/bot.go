package bot

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/luckfunc/golangBot/internal/handlers"
	"github.com/luckfunc/golangBot/internal/services"
)

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

func Run() error {
	bot := openwechat.DefaultBot(openwechat.Desktop)

	// Register QR code callback
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// Create hot reload storage object
	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()

	// Perform hot login
	if err := performHotLogin(bot, reloadStorage); err != nil {
		fmt.Println(err)
		return nil
	}

	// Handle group messages
	bot.MessageHandler = handlers.HandleGroupMessage
	services.StartDailyWatchlistPush(bot)
	services.StartIntervalWatchlistPush(bot)

	// Block until exit
	bot.Block()
	return nil
}

package main

import (
	"fmt"
	"github.com/luckfunc/golangBot/internal/bot"
)

func main() {
	if err := bot.Run(); err != nil {
		fmt.Printf("Error running bot: %v\n", err)
	}
}

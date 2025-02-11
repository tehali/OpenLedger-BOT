package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"openledger/internal/bot"
)

func main() {
	// 创建一个新的bot实例
	b := bot.NewOpenLedger()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动bot
	go func() {
		if err := b.Start(); err != nil {
			fmt.Printf("Bot error: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println("Bot is running. Press Ctrl+C to exit...")
	// 等待中断信号
	<-sigChan
	fmt.Println("\nShutting down...")
	b.Stop()
} 
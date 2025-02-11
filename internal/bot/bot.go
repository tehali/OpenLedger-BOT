package bot

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/google/uuid"
)

type OpenLedger struct {
	extensionID string
	proxies     []string
	proxyIndex  int
	proxyMutex  sync.Mutex
	running     bool
	wg          sync.WaitGroup
	logger      *Logger
}

func NewOpenLedger() *OpenLedger {
	logger, err := NewLogger()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	return &OpenLedger{
		extensionID: "chrome-extension://ekbbplmjjgoobhdlffmgeokalelnmjjc",
		proxies:     make([]string, 0),
		proxyIndex:  0,
		running:     false,
		logger:      logger,
	}
}

func (o *OpenLedger) Start() error {
	o.running = true

	// 清理终端
	o.clearTerminal()

	// 显示欢迎信息
	o.welcome()

	o.log(color.GreenString("Starting OpenLedger Bot..."))
	o.printDivider()

	// 获取代理选项
	proxyChoice, err := o.getProxyChoice()
	if err != nil {
		return fmt.Errorf("failed to get proxy choice: %w", err)
	}

	o.log(color.YellowString("Loading configuration..."))

	// 根据选择加载代理
	if proxyChoice == 1 {
		if err := o.loadAutoProxies(); err != nil {
			return fmt.Errorf("failed to load auto proxies: %w", err)
		}
	} else if proxyChoice == 2 {
		if err := o.loadManualProxies(); err != nil {
			return fmt.Errorf("failed to load manual proxies: %w", err)
		}
	}

	// 读取账号
	accounts, err := o.loadAccounts()
	if err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	o.log(color.GreenString("Account's Total: ") + color.WhiteString("%d", len(accounts)))
	o.printDivider()

	o.log(color.GreenString("Starting all processes..."))

	// 创建一个通道来等待所有账号处理完成
	done := make(chan struct{})

	// 为每个账号启动处理
	for _, account := range accounts {
		o.wg.Add(1)
		go o.processAccount(account, proxyChoice > 0)
	}

	// 在新的goroutine中等待所有账号处理完成
	go func() {
		o.wg.Wait()
		close(done)
	}()

	// 等待所有账号处理完成
	<-done
	return nil
}

func (o *OpenLedger) Stop() {
	o.running = false
	o.wg.Wait()
	if o.logger != nil {
		o.logger.Close()
	}
}

// 实现各种辅助方法
func (o *OpenLedger) clearTerminal() {
	fmt.Print("\033[H\033[2J")
}

func (o *OpenLedger) welcome() {
	green := color.New(color.FgGreen, color.Bold)
	blue := color.New(color.FgBlue, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)

	fmt.Println()
	green.Print("Auto Ping ")
	blue.Println("Open Ledger - BOT")
	fmt.Println()
	green.Print("Rey? ")
	yellow.Println("<INI WATERMARK>")
	fmt.Println()
}

func (o *OpenLedger) log(message string) {
	if o.logger != nil {
		o.logger.Log(message)
	}
}

func (o *OpenLedger) generateID() string {
	return uuid.New().String()
}

func (o *OpenLedger) generateWorkerID(account string) string {
	return base64.StdEncoding.EncodeToString([]byte(account))
}

func (o *OpenLedger) hideAccount(account string) string {
	if len(account) <= 12 {
		return account
	}
	return account[:6] + "******" + account[len(account)-6:]
}

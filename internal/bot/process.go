package bot

import (
	"fmt"

	"github.com/fatih/color"
)

// processAccount 处理单个账号
func (o *OpenLedger) processAccount(account string, useProxy bool) {
	defer o.wg.Done()

	o.log(fmt.Sprintf("%s Starting process for account: %s",
		color.CyanString("["),
		color.WhiteString(o.hideAccount(account))))

	var proxy string
	if useProxy {
		proxy = o.getNextProxy()
		o.log(fmt.Sprintf("%s Using proxy: %s for account: %s",
			color.CyanString("["),
			color.YellowString(proxy),
			color.WhiteString(o.hideAccount(account))))
	}

	// 生成初始token
	token, err := o.generateToken(account, proxy)
	if err != nil {
		o.log(fmt.Sprintf("%s Account %s - Failed to generate initial token: %v",
			color.RedString("✗"),
			color.WhiteString(o.hideAccount(account)),
			err))
		return
	}

	o.log(fmt.Sprintf("%s Account %s - Token generated successfully",
		color.GreenString("✓"),
		color.WhiteString(o.hideAccount(account))))

	// 创建错误通道
	errChan := make(chan error, 4)
	defer close(errChan)

	// 创建done通道用于控制goroutine退出
	doneChan := make(chan struct{})
	defer close(doneChan)

	// 启动各个功能的goroutine
	go o.ProcessUserEarning(account, token, proxy, errChan)
	go o.processCheckin(account, token, proxy, errChan)
	go o.processClaimTier(account, token, proxy, errChan)
	go o.processWebSocket(account, token, useProxy, proxy, errChan)

	o.log(fmt.Sprintf("%s Account %s - All processes started",
		color.GreenString("✓"),
		color.WhiteString(o.hideAccount(account))))

	// 监听错误
	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				return
			}
			if err != nil {
				o.log(fmt.Sprintf("%s Account %s - Error: %v",
					color.RedString("✗"),
					color.WhiteString(o.hideAccount(account)),
					err))
			}
		case <-doneChan:
			return
		}
	}
}

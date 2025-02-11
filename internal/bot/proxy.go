package bot

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
)

// getProxyChoice 获取用户代理选择
func (o *OpenLedger) getProxyChoice() (int, error) {
	for {
		fmt.Println("1. Run With Auto Proxy")
		fmt.Println("2. Run With Manual Proxy") 
		fmt.Println("3. Run Without Proxy")
		fmt.Print("Choose [1/2/3] -> ")

		var choice int
		var input string
		fmt.Scanln(&input)
		_, err := fmt.Sscanf(input, "%d", &choice)
		if err != nil {
			fmt.Printf("%s Please enter a number (1, 2 or 3).\n", 
				color.RedString("Invalid input."))
			continue
		}

		if choice >= 1 && choice <= 3 {
			fmt.Printf("%s Run %s Selected.\n",
				color.GreenString("✓"),
				color.WhiteString(o.getProxyTypeString(choice)))
			return choice, nil
		}

		fmt.Printf("%s Please enter either 1, 2 or 3.\n",
			color.RedString("Invalid choice."))
	}
}

// getProxyTypeString 获取代理类型字符串
func (o *OpenLedger) getProxyTypeString(choice int) string {
	switch choice {
	case 1:
		return "With Auto Proxy"
	case 2:
		return "With Manual Proxy"
	default:
		return "Without Proxy"
	}
}

// loadAutoProxies 从网络加载代理列表
func (o *OpenLedger) loadAutoProxies() error {
	url := "https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/all.txt"
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download proxies: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 保存到文件
	if err := os.WriteFile("proxy.txt", body, 0644); err != nil {
		return fmt.Errorf("failed to write proxy file: %w", err)
	}

	// 解析代理列表
	proxies := strings.Split(string(body), "\n")
	o.proxies = make([]string, 0)
	for _, proxy := range proxies {
		if proxy = strings.TrimSpace(proxy); proxy != "" {
			o.proxies = append(o.proxies, proxy)
		}
	}

	if len(o.proxies) == 0 {
		o.log(color.RedString("No proxies found in the downloaded list!"))
		return nil
	}

	o.log(color.GreenString("Proxies successfully downloaded."))
	o.log(color.YellowString("Loaded %d proxies.", len(o.proxies)))
	o.printDivider()

	return nil
}

// loadManualProxies 从本地文件加载代理列表
func (o *OpenLedger) loadManualProxies() error {
	file, err := os.Open("manual_proxy.txt")
	if err != nil {
		return fmt.Errorf("failed to open manual_proxy.txt: %w", err)
	}
	defer file.Close()

	o.proxies = make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if proxy := strings.TrimSpace(scanner.Text()); proxy != "" {
			o.proxies = append(o.proxies, proxy)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading manual_proxy.txt: %w", err)
	}

	o.log(color.YellowString("Loaded %d proxies.", len(o.proxies)))
	o.printDivider()

	return nil
}

// printDivider 打印分隔线
func (o *OpenLedger) printDivider() {
	o.log(color.CyanString(strings.Repeat("-", 75)))
}

// getNextProxy 获取下一个代理
func (o *OpenLedger) getNextProxy() string {
	o.proxyMutex.Lock()
	defer o.proxyMutex.Unlock()

	if len(o.proxies) == 0 {
		o.log(color.RedString("No proxies available!"))
		return ""
	}

	proxy := o.proxies[o.proxyIndex]
	o.proxyIndex = (o.proxyIndex + 1) % len(o.proxies)
	return o.checkProxySchemes(proxy)
}

// checkProxySchemes 检查并添加代理协议
func (o *OpenLedger) checkProxySchemes(proxy string) string {
	schemes := []string{"http://", "https://", "socks4://", "socks5://"}
	for _, scheme := range schemes {
		if strings.HasPrefix(proxy, scheme) {
			return proxy
		}
	}
	return "http://" + proxy
} 
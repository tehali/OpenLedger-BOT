package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
)

type tokenRequest struct {
	Address string `json:"address"`
}

// generateToken 生成访问令牌
func (o *OpenLedger) generateToken(account string, proxy string) (string, error) {
	maxRetries := 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		url := "https://apitn.openledger.xyz/api/v1/auth/generate_token"
		data := tokenRequest{
			Address: account,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal token request: %w", err)
		}

		// 创建请求
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		// 设置请求头
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://testnet.openledger.xyz")
		req.Header.Set("Referer", "https://testnet.openledger.xyz/")
		req.Header.Set("User-Agent", o.generateUserAgent())

		// 创建客户端
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// 如果有代理,设置代理
		if proxy != "" {
			proxyURL, err := o.getProxyClient(proxy)
			if err != nil {
				return "", fmt.Errorf("failed to set proxy: %w", err)
			}
			client.Transport = proxyURL
		}

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				o.log(fmt.Sprintf("%s Account %s - Retrying token generation (attempt %d/%d)...",
					color.YellowString("!"),
					color.WhiteString(o.hideAccount(account)),
					attempt+1,
					maxRetries))
				time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
				continue
			}
			return "", fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// 解析响应
		var tokenResp TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			if attempt < maxRetries-1 {
				o.log(fmt.Sprintf("%s Account %s - Invalid token response, retrying (attempt %d/%d)...",
					color.YellowString("!"),
					color.WhiteString(o.hideAccount(account)),
					attempt+1,
					maxRetries))
				time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
				continue
			}
			return "", fmt.Errorf("failed to decode response: %w", err)
		}

		if tokenResp.Data.Token == "" {
			if attempt < maxRetries-1 {
				o.log(fmt.Sprintf("%s Account %s - Empty token received, retrying (attempt %d/%d)...",
					color.YellowString("!"),
					color.WhiteString(o.hideAccount(account)),
					attempt+1,
					maxRetries))
				time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
				continue
			}
			return "", fmt.Errorf("received empty token")
		}

		return tokenResp.Data.Token, nil
	}

	return "", fmt.Errorf("failed to generate token after %d attempts", maxRetries)
}

// renewToken 更新访问令牌
func (o *OpenLedger) renewToken(account string, proxy string) (string, error) {
	token, err := o.generateToken(account, proxy)
	if err != nil {
		o.log(fmt.Sprintf("%s Account %s - Failed to Renew Access Token",
			color.RedString("✗"),
			color.WhiteString(o.hideAccount(account))))
		return "", err
	}

	o.log(fmt.Sprintf("%s Account %s - Access Token Has Been Renewed",
		color.GreenString("✓"),
		color.WhiteString(o.hideAccount(account))))

	return token, nil
}

// generateUserAgent 生成随机User-Agent
func (o *OpenLedger) generateUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
}

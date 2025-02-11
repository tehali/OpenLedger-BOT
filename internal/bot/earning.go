package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// getUserReward 获取用户奖励
func (o *OpenLedger) getUserReward(account, token, proxy string) (float64, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/reward"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	if proxy != "" {
		if transport, err := o.getProxyClient(proxy); err == nil {
			client.Transport = transport
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		newToken, err := o.renewToken(account, proxy)
		if err != nil {
			return 0, fmt.Errorf("token renewal failed: %w", err)
		}
		return o.getUserReward(account, newToken, proxy)
	}

	var result UserRewardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	reward, err := strconv.ParseFloat(result.Data.TotalPoint, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse reward: %w", err)
	}

	return reward, nil
}

// getWorkerReward 获取工作者奖励
func (o *OpenLedger) getWorkerReward(account, token, proxy string) (float64, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/worker_reward"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if proxy != "" {
		if transport, err := o.getProxyClient(proxy); err == nil {
			client.Transport = transport
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		newToken, err := o.renewToken(account, proxy)
		if err != nil {
			return 0, fmt.Errorf("token renewal failed: %w", err)
		}
		return o.getWorkerReward(account, newToken, proxy)
	}

	var result WorkerRewardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return 0, nil
	}

	// 处理空字符串情况
	if result.Data[0].HeartbeatCount == "" {
		result.Data[0].HeartbeatCount = "0"
	}

	heartbeat, err := strconv.ParseFloat(result.Data[0].HeartbeatCount, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse heartbeat count: %w", err)
	}

	return heartbeat, nil
}

// ProcessUserEarning 处理用户收益查询
func (o *OpenLedger) ProcessUserEarning(account, token, proxy string, errChan chan<- error) {
	for o.running {
		reward := float64(0) // 基础奖励
		heartbeat_today := float64(0)

		// 获取基础奖励（总分）
		if userReward, err := o.getUserReward(account, token, proxy); err == nil {
			reward = userReward
		}

		// 获取今日实时奖励
		if realtimeReward, err := o.getRealtimeReward(account, token, proxy); err == nil {
			heartbeat_today = realtimeReward
		}

		totalPoint := reward + heartbeat_today // 总分 = 基础奖励 + 今日奖励

		o.log(fmt.Sprintf("%s Account: %s - Earning: Total %.2f PTS - Today %.2f PTS",
			color.CyanString("["),
			color.WhiteString(o.hideAccount(account)),
			totalPoint,
			heartbeat_today))

		// 每10分钟查询一次
		time.Sleep(10 * time.Minute)
	}
}

// getRealtimeReward 获取实时奖励
func (o *OpenLedger) getRealtimeReward(account, token, proxy string) (float64, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/reward_realtime"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if proxy != "" {
		if transport, err := o.getProxyClient(proxy); err == nil {
			client.Transport = transport
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		newToken, err := o.renewToken(account, proxy)
		if err != nil {
			return 0, fmt.Errorf("token renewal failed: %w", err)
		}
		return o.getRealtimeReward(account, newToken, proxy)
	}

	var result RealtimeRewardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		if strings.Contains(err.Error(), "invalid character") {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Data) == 0 {
		return 0, nil
	}

	if result.Data[0].TotalHeartbeats == "" {
		return 0, nil
	}

	reward, err := strconv.ParseFloat(result.Data[0].TotalHeartbeats, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse reward: %w", err)
	}

	return reward, nil
}

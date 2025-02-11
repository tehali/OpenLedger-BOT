package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
)

// getCheckinDetails 获取签到详情
func (o *OpenLedger) getCheckinDetails(account, token, proxy string) (*CheckinDetailsResponse, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/claim_details"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		newToken, err := o.renewToken(account, proxy)
		if err != nil {
			return nil, fmt.Errorf("token renewal failed: %w", err)
		}
		return o.getCheckinDetails(account, newToken, proxy)
	}

	var result CheckinDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// claimCheckin 领取签到奖励
func (o *OpenLedger) claimCheckin(account, token, proxy string) (*ClaimCheckinResponse, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/claim_reward"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		newToken, err := o.renewToken(account, proxy)
		if err != nil {
			return nil, fmt.Errorf("token renewal failed: %w", err)
		}
		return o.claimCheckin(account, newToken, proxy)
	}

	var result ClaimCheckinResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// processCheckin 处理签到
func (o *OpenLedger) processCheckin(account, token, proxy string, errChan chan<- error) {
	for o.running {
		details, err := o.getCheckinDetails(account, token, proxy)
		if err != nil {
			errChan <- fmt.Errorf("get checkin details failed: %w", err)
			time.Sleep(time.Minute)
			continue
		}

		if !details.Data.Claimed {
			claim, err := o.claimCheckin(account, token, proxy)
			if err != nil {
				errChan <- fmt.Errorf("claim checkin failed: %w", err)
				time.Sleep(time.Minute)
				continue
			}

			if claim.Data.Claimed {
				o.log(fmt.Sprintf("%s Account: %s - Check-In: Is Claimed - Reward: %.2f PTS",
					color.CyanString("["),
					color.WhiteString(o.hideAccount(account)),
					details.Data.DailyPoint))
			} else {
				o.log(fmt.Sprintf("%s Account: %s - Check-In: Isn't Claimed",
					color.CyanString("["),
					color.WhiteString(o.hideAccount(account))))
			}
		} else {
			o.log(fmt.Sprintf("%s Account: %s - Check-In: Is Already Claimed",
				color.CyanString("["),
				color.WhiteString(o.hideAccount(account))))
		}

		// 每24小时检查一次
		time.Sleep(24 * time.Hour)
	}
}

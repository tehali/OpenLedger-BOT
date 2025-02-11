package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
)

type claimTierRequest struct {
	TierID int `json:"tierId"`
}

// getTierDetails 获取等级详情
func (o *OpenLedger) getTierDetails(account, token, proxy string) (*TierDetailsResponse, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/tier_details"

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
		return o.getTierDetails(account, newToken, proxy)
	}

	var result TierDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// claimTier 领取等级奖励
func (o *OpenLedger) claimTier(account, token, proxy string, tierID int) (*ClaimTierResponse, error) {
	url := "https://rewardstn.openledger.xyz/api/v1/claim_tier"

	data := claimTierRequest{
		TierID: tierID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
		return o.claimTier(account, newToken, proxy, tierID)
	} else if resp.StatusCode == 420 {
		return nil, nil
	}

	var result ClaimTierResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// processClaimTier 处理等级奖励领取
func (o *OpenLedger) processClaimTier(account, token, proxy string, errChan chan<- error) {
	for o.running {
		tiers, err := o.getTierDetails(account, token, proxy)
		if err != nil {
			errChan <- fmt.Errorf("get tier details failed: %w", err)
			time.Sleep(time.Minute)
			continue
		}

		if tiers == nil || len(tiers.Data.TierDetails) == 0 {
			o.log(fmt.Sprintf("%s Account: %s - Tier: GET Data Failed",
				color.CyanString("["),
				color.WhiteString(o.hideAccount(account))))
			time.Sleep(24 * time.Hour)
			continue
		}

		completed := true
		for _, tier := range tiers.Data.TierDetails {
			if !tier.ClaimStatus {
				completed = false
				claim, err := o.claimTier(account, token, proxy, tier.ID)
				if err != nil {
					errChan <- fmt.Errorf("claim tier failed: %w", err)
					continue
				}

				if claim != nil && claim.Status == "SUCCESS" {
					o.log(fmt.Sprintf("%s Account: %s - Tier: %s - Status: Is Claimed - Reward: %.2f PTS",
						color.CyanString("["),
						color.WhiteString(o.hideAccount(account)),
						tier.Name,
						tier.Value))
				} else {
					o.log(fmt.Sprintf("%s Account: %s - Tier: %s - Status: Not Eligible to Claim",
						color.CyanString("["),
						color.WhiteString(o.hideAccount(account)),
						tier.Name))
				}
				time.Sleep(time.Second)
			}
		}

		if completed {
			o.log(fmt.Sprintf("%s Account: %s - Tier: All Available Tier Is Completed",
				color.CyanString("["),
				color.WhiteString(o.hideAccount(account))))
		}

		// 每24小时检查一次
		time.Sleep(24 * time.Hour)
	}
}

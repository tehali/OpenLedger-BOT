package bot

// UserRewardResponse 用户奖励响应
type UserRewardResponse struct {
	Data struct {
		TotalPoint string `json:"totalPoint"`
	} `json:"data"`
}

// WorkerRewardResponse 工作者奖励响应
type WorkerRewardResponse struct {
	Data []struct {
		HeartbeatCount string `json:"heartbeat_count"`
		TotalHeartbeats string `json:"total_heartbeats"`
	} `json:"data"`
}

// RealtimeRewardResponse 实时奖励响应
type RealtimeRewardResponse struct {
	Data []struct {
		TotalHeartbeats string `json:"total_heartbeats"`
	} `json:"data"`
}

// CheckinDetailsResponse 签到详情响应
type CheckinDetailsResponse struct {
	Data struct {
		Claimed    bool    `json:"claimed"`
		DailyPoint float64 `json:"dailyPoint"`
	} `json:"data"`
}

// ClaimCheckinResponse 签到领取响应
type ClaimCheckinResponse struct {
	Data struct {
		Claimed bool `json:"claimed"`
	} `json:"data"`
}

// TierDetailsResponse 等级详情响应
type TierDetailsResponse struct {
	Data struct {
		TierDetails []TierDetail `json:"tierDetails"`
	} `json:"data"`
}

type TierDetail struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	ClaimStatus bool    `json:"claimStatus"`
}

// ClaimTierResponse 等级奖励领取响应
type ClaimTierResponse struct {
	Status string `json:"status"`
} 
package bot

type TokenResponse struct {
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type WorkerMessage struct {
	WorkerID   string      `json:"workerID"`
	MsgType    string      `json:"msgType"`
	WorkerType string      `json:"workerType"`
	Message    interface{} `json:"message"`
}

type RegisterMessage struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Worker  Worker `json:"worker"`
}

type Worker struct {
	Host         string `json:"host"`
	Identity     string `json:"identity"`
	OwnerAddress string `json:"ownerAddress"`
	Type         string `json:"type"`
}

type HeartbeatMessage struct {
	Worker   Worker   `json:"Worker"`
	Capacity Capacity `json:"Capacity"`
}

type Capacity struct {
	AvailableMemory  float64  `json:"AvailableMemory"`
	AvailableStorage string   `json:"AvailableStorage"`
	AvailableGPU     string   `json:"AvailableGPU"`
	AvailableModels  []string `json:"AvailableModels"`
} 
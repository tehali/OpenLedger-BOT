package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

// WebSocket消息类型
const (
	MsgTypeRegister  = "REGISTER"
	MsgTypeHeartbeat = "HEARTBEAT"
	MsgTypeJob       = "JOB"
	MsgTypeResponse  = "RESPONSE"
)

// 添加心跳响应消息类型
type HeartbeatResponse struct {
	Status bool   `json:"status"`
	Type   string `json:"type"`
}

// connectWebSocket 建立WebSocket连接
func (o *OpenLedger) connectWebSocket(account, token string, proxy string) (*websocket.Conn, error) {
	// 构建WebSocket URL
	wsURL := fmt.Sprintf("wss://apitn.openledger.xyz/ws/v1/orch?authToken=%s", token)
	o.log(fmt.Sprintf("Connecting WebSocket for account %s", o.hideAccount(account)))

	// 设置请求头
	headers := http.Header{
		"Accept-Encoding": {"gzip, deflate, br, zstd"},
		"Accept-Language": {"id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7"},
		"Cache-Control":   {"no-cache"},
		"Origin":          {o.extensionID},
		"Pragma":          {"no-cache"},
		"User-Agent":      {o.generateUserAgent()},
	}

	// 创建Dialer
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	// 如果有代理,设置代理
	if proxy != "" {
		if transport, err := o.getProxyClient(proxy); err == nil {
			dialer.NetDial = transport.Dial
		}
	}

	// 连接WebSocket
	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect websocket: %w", err)
	}

	return conn, nil
}

// sendRegisterMessage 发送注册消息
func (o *OpenLedger) sendRegisterMessage(conn *websocket.Conn, account string) error {
	id := o.generateID()
	identity := o.generateWorkerID(account)

	msg := WorkerMessage{
		WorkerID:   identity,
		MsgType:    MsgTypeRegister,
		WorkerType: "LWEXT",
		Message: RegisterMessage{
			ID:   id,
			Type: MsgTypeRegister,
			Worker: Worker{
				Host:         o.extensionID,
				Identity:     identity,
				OwnerAddress: account,
				Type:         "LWEXT",
			},
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send register message: %w", err)
	}

	return nil
}

// sendHeartbeatMessage 发送心跳消息
func (o *OpenLedger) sendHeartbeatMessage(conn *websocket.Conn, account string) error {
	identity := o.generateWorkerID(account)
	memory := 32.0
	storage := "500.00"

	msg := WorkerMessage{
		WorkerID:   identity,
		MsgType:    MsgTypeHeartbeat,
		WorkerType: "LWEXT",
		Message: HeartbeatMessage{
			Worker: Worker{
				Identity:     identity,
				OwnerAddress: account,
				Type:         "LWEXT",
				Host:         o.extensionID,
			},
			Capacity: Capacity{
				AvailableMemory:  memory,
				AvailableStorage: storage,
				AvailableGPU:     "",
				AvailableModels:  []string{},
			},
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send heartbeat message: %w", err)
	}

	o.log(fmt.Sprintf("%s Account %s - Heartbeat sent",
		color.CyanString("["),
		color.WhiteString(o.hideAccount(account))))

	return nil
}

// handleWebSocketMessage 处理WebSocket消息
func (o *OpenLedger) handleWebSocketMessage(conn *websocket.Conn, account string) error {
	_, message, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return fmt.Errorf("websocket read error: %w", err)
		}
		return err
	}

	var msg map[string]interface{}
	// 解析消息
	if err := json.Unmarshal(message, &msg); err != nil {
		o.log(fmt.Sprintf("%s Account %s - Failed to parse message: %v",
			color.YellowString("!"),
			color.WhiteString(o.hideAccount(account)),
			err))
		return nil
	}

	msgType, ok := msg["msgType"].(string)
	if !ok {
		// 尝试其他字段
		msgType, ok = msg["type"].(string)
		if !ok {
			return nil
		}
	}

	switch msgType {
	case MsgTypeRegister:
		o.log(fmt.Sprintf("%s Account %s - WebSocket registered successfully",
			color.GreenString("✓"),
			color.WhiteString(o.hideAccount(account))))
		return nil

	case MsgTypeHeartbeat:
		if message, ok := msg["message"].(map[string]interface{}); ok {
			if status, ok := message["Status"].(bool); ok && status {
				o.log(fmt.Sprintf("%s Account %s - Heartbeat acknowledged",
					color.GreenString("✓"),
					color.WhiteString(o.hideAccount(account))))
			}
		}
		return nil

	case MsgTypeJob:
		response := map[string]interface{}{
			"workerID":   o.generateWorkerID(account),
			"msgType":    "JOB_ASSIGNED",
			"workerType": "LWEXT",
			"message": map[string]interface{}{
				"Status": true,
				"Ref":    msg["UUID"],
			},
		}
		if err := conn.WriteJSON(response); err != nil {
			return fmt.Errorf("failed to send job response: %w", err)
		}
		o.log(fmt.Sprintf("%s Account %s - Job assigned",
			color.GreenString("✓"),
			color.WhiteString(o.hideAccount(account))))

	case MsgTypeResponse:
		return nil

	default:
		o.log(fmt.Sprintf("%s Account %s - Unknown message type: %s",
			color.YellowString("!"),
			color.WhiteString(o.hideAccount(account)),
			msgType))
	}

	return nil
}

// processWebSocket 处理WebSocket连接
func (o *OpenLedger) processWebSocket(account, token string, useProxy bool, proxy string, errChan chan<- error) {
	reconnectDelay := time.Second * 5
	for o.running {
		retries := 0
		maxRetries := 3
		// 建立连接
		actualProxy := ""
		if useProxy {
			actualProxy = proxy
		}
		conn, err := o.connectWebSocket(account, token, actualProxy)
		if err != nil {
			errChan <- fmt.Errorf("websocket connection failed: %w", err)
			if retries < maxRetries {
				retries++
				time.Sleep(time.Duration(retries) * 5 * time.Second)
				continue
			}
			time.Sleep(30 * time.Second)
			retries = 0
			continue
		}

		// 连接成功后输出状态
		identity := o.generateWorkerID(account)
		o.log(fmt.Sprintf("%s Account: %s - Proxy: %s - Worker ID: %s - Status: %s",
			color.CyanString("["),
			color.WhiteString(o.hideAccount(account)),
			color.WhiteString(actualProxy),
			color.WhiteString(o.hideAccount(identity)),
			color.GreenString("Webscoket Is Connected")))

		// 发送注册消息
		if err := o.sendRegisterMessage(conn, account); err != nil {
			errChan <- fmt.Errorf("register message failed: %w", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		// 启动心跳goroutine
		heartbeatTicker := time.NewTicker(30 * time.Second)
		go func() {
			for range heartbeatTicker.C {
				if err := o.sendHeartbeatMessage(conn, account); err != nil {
					errChan <- fmt.Errorf("heartbeat message failed: %w", err)
					return
				}
			}
		}()

		// 处理消息
		for o.running {
			if err := o.handleWebSocketMessage(conn, account); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					errChan <- fmt.Errorf("websocket error: %w", err)
				}
				break
			}
		}

		// 清理
		heartbeatTicker.Stop()
		conn.Close()

		// 连接断开后输出状态
		o.log(fmt.Sprintf("%s Account: %s - Proxy: %s - Worker ID: %s - Status: %s",
			color.CyanString("["),
			color.WhiteString(o.hideAccount(account)),
			color.WhiteString(actualProxy),
			color.WhiteString(o.hideAccount(identity)),
			color.YellowString("Webscoket Connection Closed")))

		// 如果程序还在运行,等待后重试
		if o.running {
			time.Sleep(reconnectDelay)
			// 增加重连延迟，最大30秒
			if reconnectDelay < time.Second*30 {
				reconnectDelay *= 2
			}
		}
	}
}

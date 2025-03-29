package volc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/domain/model"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var _ model.TtsApp = (*VcTtsApp)(nil)

// VcTtsApp 是火山引擎的大模型语音合成
// 默认双向流式, 暂定一次对话共用一个连接, 如果太长了之后就一轮话一个连接
// 默认使用PCM格式, 24000采样频率
type VcTtsApp struct {
	// ws 连接
	ws *websocket.Conn
	mu sync.Mutex

	appKey     string
	accessKey  string
	speaker    string
	resourceId string
	url        string

	// connId 连接id, 标识一次连接
	connId string
	// logId 服务端返回的logId, 用于定位问题
	logId string
	// sessionId
	sessionId string
	// header 是请求头, 携带鉴权信息
	header http.Header
}

func NewVcTtsApp(appKey, accessKey, speaker, resourceId, url string) *VcTtsApp {
	connId := uuid.New().String()
	logId := genLogID()
	sessionId := uuid.New().String()
	app := &VcTtsApp{
		ws:         nil,
		appKey:     appKey,
		accessKey:  accessKey,
		speaker:    speaker,
		url:        url,
		resourceId: resourceId,
		connId:     connId,
		logId:      logId,
		sessionId:  sessionId,
		mu:         sync.Mutex{},
	}
	app.buildHTTPHeader()
	return app
}

// Dial 建立ws连接
func (app *VcTtsApp) Dial() error {
	conn, r, err := websocket.DefaultDialer.DialContext(context.Background(), app.url, app.header)
	if err != nil {
		if r != nil {
			body, parseErr := io.ReadAll(r.Body)
			if parseErr != nil {
				parseErr = fmt.Errorf("parse response body failed: %w", parseErr)
				body = []byte(parseErr.Error())
			}
			err = fmt.Errorf("[code=%s] [body=%s] %w", r.Status, body, err)
		}
	}
	app.ws = conn
	return err
}

// Start 应用层协议握手
func (app *VcTtsApp) Start() (err error) {
	if err = app.startConnection(); err != nil {
		return
	}
	namespace := "BidirectionalTTS"
	// TODO: 之后可能需要指定采样频率
	params := &TTSReqParams{
		Speaker: app.speaker,
		AudioParams: &AudioParams{
			Format:     "pcm",
			SampleRate: 24000,
		},
	}
	if err = app.startTTSSession(namespace, params); err != nil {
		return
	}
	return
}

// startConnection 建立application级别的连接
func (app *VcTtsApp) startConnection() error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create StartSession request message: %w", err)
	}
	msg.Event = int32(EventStartConnection)
	msg.Payload = []byte("{}")

	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal StartConnection request message: %w", err)
	}

	if err := app.ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send StartConnection request: %w", err)
	}

	// Read ConnectionStarted message.
	mt, frame, err := app.ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("read ConnectionStarted response: %w", err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return fmt.Errorf("unexpected Websocket message type: %d", mt)
	}

	msg, _, err = Unmarshal(frame, protocol.ContainsSequence)
	if err != nil {
		glog.Infof("StartConnection response: %s", frame)
		return fmt.Errorf("unmarshal ConnectionStarted response message: %w", err)
	}
	if msg.Type != MsgTypeFullServer {
		return fmt.Errorf("unexpected ConnectionStarted message type: %s", msg.Type)
	}
	if Event(msg.Event) != EventConnectionStarted {
		return fmt.Errorf("unexpected response event (%s) for StartConnection request", Event(msg.Event))
	}
	glog.Infof("Connection started (event=%s) connectID: %s, payload: %s", Event(msg.Event), msg.ConnectID, msg.Payload)

	return nil
}

// startTTSSession 开启TTSSession, 应该是用于标识一段上下文
func (app *VcTtsApp) startTTSSession(namespace string, params *TTSReqParams) error {
	req := TTSRequest{
		Event:     int32(EventStartSession),
		Namespace: namespace,
		ReqParams: params,
	}
	payload, err := json.Marshal(&req)
	glog.Infof("StartSession request payload: %s", string(payload))
	if err != nil {
		return fmt.Errorf("marshal StartSession request payload: %w", err)
	}

	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create StartSession request message: %w", err)
	}
	msg.Event = req.Event
	msg.SessionID = app.sessionId
	msg.Payload = payload

	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal StartSession request message: %w", err)
	}

	if err := app.ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send StartSession request: %w", err)
	}

	// Read SessionStarted message.
	mt, frame, err := app.ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("read SessionStarted response: %w", err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return fmt.Errorf("unexpected Websocket message type: %d", mt)
	}

	// Validate SessionStarted message.
	msg, _, err = Unmarshal(frame, protocol.ContainsSequence)
	if err != nil {
		glog.Infof("StartSession response: %s", frame)
		return fmt.Errorf("unmarshal SessionStarted response message: %w", err)
	}
	if msg.Type != MsgTypeFullServer {
		return fmt.Errorf("unexpected SessionStarted message type: %s", msg.Type)
	}
	if Event(msg.Event) != EventSessionStarted {
		return fmt.Errorf("unexpected response event (%s) for StartSession request", Event(msg.Event))
	}
	glog.Infof("%s session started with ID: %s", namespace, msg.SessionID)

	return nil
}

// Send 发送请求
func (app *VcTtsApp) Send(text string) (err error) {
	return app.sendTtsMessage(text)
}

// sendTtsMessage 发送一条tts消息
func (app *VcTtsApp) sendTtsMessage(text string) error {
	req := TTSRequest{
		Event:     int32(EventTaskRequest),
		Namespace: "BidirectionalTTS",
		ReqParams: &TTSReqParams{
			Text:    text,
			Speaker: app.speaker,
			AudioParams: &AudioParams{
				Format:     "pcm",
				SampleRate: 24000,
			},
		},
	}
	payload, err := json.Marshal(&req)
	if err != nil {
		return fmt.Errorf("marshal TaskRequest request payload: %w", err)
	}

	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create TaskRequest request message: %w", err)
	}
	msg.Event = req.Event
	msg.SessionID = app.sessionId
	msg.Payload = payload

	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal TaskRequest request message: %w", err)
	}

	app.mu.Lock()
	defer app.mu.Unlock()
	if err := app.ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send TaskRequest request: %w", err)
	}

	glog.Info("TaskRequest request is sent.")
	return nil
}

// Receive 接收请求
func (app *VcTtsApp) Receive() []byte {
	for {
		msg, err := app.receiveMessage()
		if err != nil {
			glog.Errorf("Receive message error: %v", err)
			return nil
		}
		switch msg.Type {
		case MsgTypeFullServer:
			glog.Infof("Receive text message (event=%s, session_id=%s): %s", Event(msg.Event), msg.SessionID, msg.Payload)
			if msg.Event == int32(EventSessionFinished) {
				log.Info("event type:", msg.Event)
				return nil
			}
			continue

		case MsgTypeAudioOnlyServer:
			glog.Infof("Receive audio message (event=%s): session_id=%s", Event(msg.Event), msg.SessionID)
			return msg.Payload

		case MsgTypeError:
			glog.Errorf("Receive Error message (code=%d): %s", msg.ErrorCode, msg.Payload)
			return nil
		default:
			glog.Errorf("Received unexpected message type: %s", msg.Type)
			return nil
		}
	}
}

// receiveMessage 从ws中接受消息
func (app *VcTtsApp) receiveMessage() (*Message, error) {
	mt, frame, err := app.ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return nil, fmt.Errorf("unexpected Websocket message type: %d", mt)
	}

	msg, _, err := Unmarshal(frame, ContainsSequence)
	if err != nil {
		if len(frame) > 500 {
			frame = frame[:500]
		}
		glog.Infof("Data response: %s", frame)
		return nil, fmt.Errorf("unmarshal response message: %w", err)
	}
	return msg, nil
}

// Close 关闭连接释放资源
func (app *VcTtsApp) Close() (err error) {
	if err = app.finishSession(); err != nil {
		glog.Errorf("Close session finished with error: %v", err)
	}
	if err = app.finishConnection(); err != nil {
		glog.Errorf("Close connection finished with error: %v", err)
	}
	return app.ws.Close()
}

// finishSession 关闭session
func (app *VcTtsApp) finishSession() error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create FinishSession request message: %w", err)
	}
	msg.Event = int32(EventFinishSession)
	msg.SessionID = app.sessionId
	msg.Payload = []byte("{}")

	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal FinishSession request message: %w", err)
	}

	if err := app.ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send FinishSession request: %w", err)
	}

	glog.Info("FinishSession request is sent.")
	return nil
}

// finishConnection 关闭连接
func (app *VcTtsApp) finishConnection() error {
	msg, err := NewMessage(MsgTypeFullClient, MsgTypeFlagWithEvent)
	if err != nil {
		return fmt.Errorf("create FinishConnection request message: %w", err)
	}
	msg.Event = int32(EventFinishConnection)
	msg.Payload = []byte("{}")

	frame, err := protocol.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal FinishConnection request message: %w", err)
	}

	if err := app.ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return fmt.Errorf("send FinishConnection request: %w", err)
	}

	// Read ConnectionStarted message.
	mt, frame, err := app.ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("read ConnectionFinished response: %w", err)
	}
	if mt != websocket.BinaryMessage && mt != websocket.TextMessage {
		return fmt.Errorf("unexpected Websocket message type: %d", mt)
	}

	msg, _, err = Unmarshal(frame, protocol.ContainsSequence)
	if err != nil {
		glog.Infof("FinishConnection response: %s", frame)
		return fmt.Errorf("unmarshal ConnectionFinished response message: %w", err)
	}
	if msg.Type != MsgTypeFullServer {
		return fmt.Errorf("unexpected ConnectionFinished message type: %s", msg.Type)
	}
	if Event(msg.Event) != EventConnectionFinished {
		return fmt.Errorf("unexpected response event (%s) for FinishConnection request", Event(msg.Event))
	}

	glog.Infof("Connection finished (event=%s)", Event(msg.Event))
	return nil
}

// protocol 是火山tts的二进制帧协议
var protocol = NewBinaryProtocol()

func init() {
	// Initialize binary protocol settings.
	protocol.SetVersion(Version1)
	protocol.SetHeaderSize(HeaderSize4)
	protocol.SetSerialization(SerializationJSON)
	protocol.SetCompression(CompressionNone, nil)
	protocol.ContainsSequence = ContainsSequence
}

// buildHTTPHeader 构造请求头
func (app *VcTtsApp) buildHTTPHeader() {
	app.header = http.Header{
		"X-Tt-Logid":        []string{app.logId},
		"X-Api-Resource-Id": []string{app.resourceId},
		"X-Api-Access-Key":  []string{app.accessKey},
		"X-Api-App-Key":     []string{app.appKey},
		"X-Api-Connect-Id":  []string{app.connId},
	}
}

// genLogID 生成日志ID
func genLogID() string {
	const (
		maxRandNum = 1<<24 - 1<<20
		length     = 53
		version    = "02"
		localIP    = "00000000000000000000000000000000"
	)
	ts := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	r := uint64(fastrand.Uint32n(maxRandNum) + 1<<20)
	var sb strings.Builder
	sb.Grow(length)
	sb.WriteString(version)
	sb.WriteString(strconv.FormatUint(ts, 10))
	sb.WriteString(localIP)
	sb.WriteString(strconv.FormatUint(r, 16))
	return sb.String()
}

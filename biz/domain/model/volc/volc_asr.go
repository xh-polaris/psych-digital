package volc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

// VcAsrApp 是火山引擎的大模型语音识别
// 默认双向流式, 暂定一次对话共用一个连接, 如果太长了就一段话一个连接
// 目前只支持单声道音频, 默认使用pcm格式, 16000采样频率, 全量返回
type VcAsrApp struct {
	// ws 连接
	ws *websocket.Conn

	appKey     string
	accessKey  string
	resourceId string
	url        string

	// seq 发送的消息序列号
	seq int
	// connId 连接id, 标识一次连接
	connId string
	// logId 服务端返回的logId, 用于定位问题
	logId string
	// sessionId
	sessionId string
	// header 是请求头, 携带鉴权信息
	header http.Header
}

// NewVcAsrApp 构造一个新的
func NewVcAsrApp(appKey, accessKey, resourceId, url string) *VcAsrApp {
	connId := uuid.New().String()
	logId := genLogID()
	sessionId := uuid.New().String()
	app := &VcAsrApp{
		ws:         nil,
		appKey:     appKey,
		accessKey:  accessKey,
		url:        url,
		resourceId: resourceId,
		connId:     connId,
		logId:      logId,
		sessionId:  sessionId,
	}
	app.buildHTTPHeader()
	return app
}

// Dial 建立ws连接
func (app *VcAsrApp) Dial() error {
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
	if r != nil {
		log.Info("X-Tt-Logid: ", r.Header.Get("X-Tt-Logid"))
	}
	app.ws = conn
	return err
}

// Start 完成应用层协议握手
func (app *VcAsrApp) Start() error {
	var err error

	// 协商配置参数
	// 用户参数
	user := map[string]string{
		"uid": app.connId,
	}
	// 音频参数 TODO: 目前格式均固定, 之后允许配置
	audio := map[string]any{
		"format":      "pcm", // 格式,  pcm/wav/ogg
		"sample_rate": 16000, // 采样频率, 只支持16000
		"bits":        16,    // 采样位数, 默认16 TODO: 确认
		"channels":    1,     // 单声道, TODO: 确认
		"codec":       "raw", // 编码方式, raw(pcm)
	}
	//  请求参数
	request := map[string]any{
		"model_name":  "bigmodel", // 目前只有这个模型
		"enable_punc": true,       // 启用标点
	}
	payload := map[string]any{
		"user":    user,
		"audio":   audio,
		"request": request,
	}

	// 序列化为字节
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// gzip压缩
	payloadBytes, err = util.GzipCompress(payloadBytes)
	if err != nil {
		return err
	}
	// 组装full client request, full client request = header + sequence + payload
	header := getHeader(FullClientRequest, PosSequence, JSON, GZIP, byte(0))
	seqBytes := intToBytes(app.seq)
	payloadSize := intToBytes(len(payloadBytes))
	fullClientRequest := append(header, append(seqBytes, append(payloadSize, payloadBytes...)...)...)
	if err = app.ws.WriteMessage(websocket.BinaryMessage, fullClientRequest); err != nil {
		return err
	}
	return nil
}

// Send 发送音频流
func (app *VcAsrApp) Send(data []byte) error {
	// 此处本来应该在最后一个包时, 将seq置为负数, 然后采用结束帧类型, 但是考虑到采用Close方法结束, 所以这里就不用这种方式了, 而是在Close中粗暴退出
	app.seq++
	messageTypeSpecificFlags := PosSequence
	// header
	header := getHeader(AudioOnlyRequest, byte(messageTypeSpecificFlags), JSON, GZIP, byte(0))
	// seq
	seqBytes := intToBytes(app.seq)
	// payload
	payloadSize := intToBytes(len(data))
	payloadBytes, err := util.GzipCompress(data)
	if err != nil {
		return err
	}

	audioOnlyRequest := append(header, append(seqBytes, append(payloadSize, payloadBytes...)...)...)
	if err = app.ws.WriteMessage(websocket.BinaryMessage, audioOnlyRequest); err != nil {
		return err
	}
	return nil
}

// Receive 接受响应
func (app *VcAsrApp) Receive() ([]byte, error) {
	mt, res, err := app.ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	switch mt {
	case websocket.BinaryMessage:
		return app.receiveText(res)
	case websocket.TextMessage:
		return app.receiveBytes(res)
	default:
		return nil, fmt.Errorf("invalid websocket message")
	}
}

// receiveText 接受到文本消息
func (app *VcAsrApp) receiveText(res []byte) ([]byte, error) {
	log.Info("receiveText: ", string(res))
	return nil, nil
}

// receiveBytes 接收到字节流
func (app *VcAsrApp) receiveBytes(res []byte) ([]byte, error) {
	data, seq, err := parse(res)
	// seq 小于0 表示这是最后一个包, 后续没有了
	if err != nil || seq < 0 {
		return nil, err
	}
	return data, err
}

// Close 释放资源
func (app *VcAsrApp) Close() error {
	return app.ws.Close()
}

// parse 解析响应帧
func parse(res []byte) (data []byte, seq int, err error) {
	if res == nil || len(res) == 0 {
		return
	}
	num := byte(0b00001111)

	// header 32bits, 大部分字段暂时没有实际作用
	//_protocolVersion := (res[0] >> 4) & num
	//_headerSize := res[0] & 0x0f
	_messageType := (res[1] >> 4) & num
	//_messageTypeSpecificFlags := res[1] & 0x0f
	//_serializationMethod := res[2] >> num
	_messageCompression := res[2] & 0x0f
	//_reserved := res[3]

	// sequence 4bytes
	_seq, err := bytesToInt(res[4:8])
	if err != nil {
		return nil, 0, err
	}

	// payload size 4 byte, 暂时没有实际作用
	//_payloadSize, err := bytesToInt(res[8:12])
	//if err != nil {
	//	return nil, 0, err
	//}

	// 有效数据
	payload := res[12:]

	switch _messageType {
	case FullClientRequest:
		if _messageCompression == GZIP {
			data, err = util.GzipDecompress(payload)
			return data, _seq, err
		} else {
			return payload, _seq, nil
		}
	case ServerAck:
		return payload, _seq, nil
	case ServerErrorResponse:
		return payload, _seq, fmt.Errorf("code: %d, msg: %s", _seq, string(payload))
	default:
		return nil, 0, fmt.Errorf("unknown message type: %d", _messageType)
	}
}

// buildHTTPHeader 构造鉴权请求头
func (app *VcAsrApp) buildHTTPHeader() {
	app.header = http.Header{
		"X-Api-Resource-Id": []string{app.resourceId},
		"X-Api-Access-Key":  []string{app.accessKey},
		"X-Api-App-Key":     []string{app.appKey},
		"X-Api-Connect-Id":  []string{app.connId},
	}
}

// getHeader 生成协议头
func getHeader(messageType, messageTypeSpecificFlags, serialMethod, compressionType, reserverData byte) []byte {
	header := make([]byte, 4)
	header[0] = (ProtocolVersion << 4) | DefaultHeaderSize
	header[1] = (messageType << 4) | messageTypeSpecificFlags
	header[2] = (serialMethod << 4) | compressionType
	header[3] = reserverData
	return header
}

// intToBytes 将整数变成字节数组
func intToBytes(n int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}

// bytesToInt 将字节数组变成整数
func bytesToInt(data []byte) (int, error) {
	if len(data) != 4 || data == nil {
		return 0, fmt.Errorf("bytesToInt err")
	}
	return int(binary.BigEndian.Uint32(data)), nil
}

// 协议常量
const (
	ProtocolVersion   = 0b0001
	DefaultHeaderSize = 0b0001

	FullClientRequest   = 0b0001
	AudioOnlyRequest    = 0b0010
	FullServerResponse  = 0b1001
	ServerAck           = 0b1011
	ServerErrorResponse = 0b1111

	NoSequence      = 0b0000 // no check sequence
	PosSequence     = 0b0001
	NegSequence     = 0b0010
	NegWithSequence = 0b0011
	NegSequence1    = 0b0011

	NoSerialization = 0b0000
	JSON            = 0b0001

	NoCompression = 0b0000
	GZIP          = 0b0001
)

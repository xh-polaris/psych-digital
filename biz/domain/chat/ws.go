package chat

import (
	"encoding/json"
	"fmt"
	"github.com/hertz-contrib/websocket"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/consts"
	"reflect"
	"sync"
)

// WsHelper 是封装Websocket协议的工具类
// 最佳实践是单协程读, 所以不需要使用读锁, 但是涉及到文字和音频的混合传输, 所以可能需要一个协程读, 另外两个协程分别处理文本和音频
type WsHelper struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func NewWsHelper(conn *websocket.Conn) *WsHelper {
	return &WsHelper{
		mu:   sync.Mutex{},
		conn: conn,
	}
}

// Read 获取消息
func (ws *WsHelper) Read() (int, []byte, error) {
	return ws.conn.ReadMessage()
}

// ReadJson 从流中获取一个Json对象， 需要传入指针
func (ws *WsHelper) ReadJson(obj any) error {
	// 读取消息
	mt, msg, err := ws.conn.ReadMessage()

	// 需要文本类型
	if mt != websocket.TextMessage {
		return fmt.Errorf("invalid message type")
	}

	if err != nil {
		log.Error("read message error:", err)
		return err
	}

	// 不同类型处理
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Ptr:
		err = json.Unmarshal(msg, obj)
	default:
		err = fmt.Errorf("obj must be a pointer")
	}

	// 解析Json
	if err != nil {
		log.Error("unmarshal message error:", err)
		return err
	}
	return nil
}

// Error 写入一个错误信息
func (ws *WsHelper) Error(errno *consts.Errno) error {
	resp := &dto.Response{
		Code: errno.Code(),
		Msg:  errno.Error(),
	}
	return ws.WriteJson(resp)
}

// WriteJson 写入一个Json对象
func (ws *WsHelper) WriteJson(obj any) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	return ws.conn.WriteMessage(websocket.TextMessage, bytes)
}

// WriteBytes 写入字节流
func (ws *WsHelper) WriteBytes(bytes []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.conn.WriteMessage(websocket.BinaryMessage, bytes)
}

// Close 关闭连接
func (ws *WsHelper) Close() error {
	return ws.conn.Close()
}

package domain

import (
	"encoding/json"
	"github.com/xh-polaris/psych-digital/biz/application/dto"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	rs "github.com/xh-polaris/psych-digital/biz/infrastructure/redis"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"sync"
)

var (
	instance *RedisHelper
	once     sync.Once
)

type RedisHelper struct {
	rs *redis.Redis
}

func GetRedisHelper() *RedisHelper {
	c := config.GetConfig()
	once.Do(func() {
		instance = &RedisHelper{
			rs: rs.NewRedis(c),
		}
	})
	return instance
}

// AddAi 添加ai对话记录
func (r *RedisHelper) AddAi(sessionId, msg string) error {
	return r.add(sessionId, "ai", msg)
}

// AddUser 添加用户对话记录
func (r *RedisHelper) AddUser(sessionId, msg string) error {
	return r.add(sessionId, "user", msg)
}

// AddSystem 添加系统对话记录
func (r *RedisHelper) AddSystem(sessionId, msg string) error {
	return r.add(sessionId, "system", msg)
}

// add 将对话记录添加到队列尾部
func (r *RedisHelper) add(sessionId, role, msg string) error {
	history := dto.ChatHistory{
		Role:    role,
		Content: msg,
	}

	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	_, err = r.rs.Rpush(sessionId, string(data))
	return err
}

// Load 获取session对应的所有对话记录
func (r *RedisHelper) Load(sessionId string) ([]*dto.ChatHistory, error) {
	// 获取所有元素
	data, err := r.rs.Lrange(sessionId, 0, -1)
	if err != nil {
		return nil, err
	}

	var history []*dto.ChatHistory
	for _, v := range data {
		var his dto.ChatHistory
		if err = json.Unmarshal([]byte(v), &his); err != nil {
			return nil, err
		}
		history = append(history, &his)
	}
	return history, nil
}

// Remove 删除Session对应的记录
func (r *RedisHelper) Remove(sessionId string) error {
	_, err := r.rs.Del(sessionId)
	return err
}

//// MockRedis 模拟 Redis 的全局存储
//var MockRedis = struct {
//	sync.Mutex
//	data map[string][]dto.ChatHistory
//}{
//	data: make(map[string][]dto.ChatHistory),
//}
//
//type MemoryRedisHelper struct{}
//
//func NewMemoryRedisHelper() *MemoryRedisHelper {
//	return &MemoryRedisHelper{}
//}
//
//// AddAi 添加 AI 消息
//func (m *MemoryRedisHelper) AddAi(sessionId, msg string) error {
//	return m.add(sessionId, "ai", msg)
//}
//
//// AddUser 添加用户消息
//func (m *MemoryRedisHelper) AddUser(sessionId, msg string) error {
//	return m.add(sessionId, "user", msg)
//}
//
//// AddSystem 添加系统消息
//func (m *MemoryRedisHelper) AddSystem(sessionId, msg string) error {
//	return m.add(sessionId, "system", msg)
//}
//
//// 通用添加方法
//func (m *MemoryRedisHelper) add(sessionId, role, msg string) error {
//	MockRedis.Lock()
//	defer MockRedis.Unlock()
//
//	history := dto.ChatHistory{
//		Role:    role,
//		Content: msg,
//	}
//
//	// 模拟 RPUSH 操作
//	MockRedis.data[sessionId] = append(MockRedis.data[sessionId], history)
//
//	// 打印操作日志
//	printOperation("ADD", sessionId, history)
//	printRedisState("STATE", sessionId, "")
//	return nil
//}
//
//// 打印操作详情
//func printOperation(op, sessionId string, data dto.ChatHistory) {
//	jsonData, _ := json.Marshal(data)
//	fmt.Printf("[%s] %s\n", op, sessionId)
//	fmt.Printf("└─ Data: %s\n", string(jsonData))
//}
//
//// 打印当前存储状态
//func printRedisState(op, sessionId, msg string) {
//	fmt.Printf("[%s] Current Redis State\n", op)
//
//	if sessionId != "" {
//		printSession(sessionId)
//		return
//	}
//
//	for sid := range MockRedis.data {
//		printSession(sid)
//	}
//	fmt.Println("━━━━━━━━━━━━━━━━━━━━")
//}
//
//// 打印单个会话状态
//func printSession(sessionId string) {
//	fmt.Printf("┏━ Session: %s\n", sessionId)
//	for i, msg := range MockRedis.data[sessionId] {
//		fmt.Printf("┃ %d. [%s] %s\n", i+1, msg.Role, msg.Content)
//	}
//	fmt.Println("┗━━━━━━━━━━━━━━━━━━")
//}

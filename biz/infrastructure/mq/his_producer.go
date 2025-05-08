package mq

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"golang.org/x/net/context"
	"math"
	"sync"
	"time"
)

// conn 采用单例模式, 复用连接
var (
	conn *amqp.Connection
	once sync.Once
	url  string
)

// getConn 获取连接单例
func getConn() *amqp.Connection {
	once.Do(func() {
		conf := config.GetConfig()
		url = conf.RabbitMQ.Url
		c, err := amqp.Dial(url)
		if err != nil {
			util.FailOnError("rabbit mq connect failed", err)
		}
		conn = c
		// 自动重连监听
		go monitor()
	})
	return conn
}

// monitor 监听健康状态并重连
func monitor() {
	for {
		reason := <-conn.NotifyClose(make(chan *amqp.Error))
		log.Info("RabbitMQ connection closed , reason: ", reason)

		retries := 0
		for {
			time.Sleep(time.Duration(math.Pow(2, float64(retries))) * time.Second)

			newConn, err := amqp.Dial(url)
			if err == nil {
				conn = newConn
				log.Info("Reconnect to RabbitMQ")
				break
			}
			retries++
			if retries > 5 {
				util.FailOnError("超过最大重连次数5", fmt.Errorf("RabbitMQ 断开连接且重连失败"))
				return
			}
		}
	}
}

var (
	producer     *HistoryProducer
	producerOnce sync.Once
)

// HistoryProducer 历史记录生产者
type HistoryProducer struct {
	mu      sync.Mutex
	conn    *amqp.Connection
	channel *amqp.Channel
}

// GetHistoryProducer 获取历史记录生产者
func GetHistoryProducer() *HistoryProducer {
	producerOnce.Do(func() {
		c := getConn()
		ch, err := c.Channel()
		if err != nil {
			util.FailOnError("create channel failed", err)
		}
		producer = &HistoryProducer{
			conn:    c,
			channel: ch,
		}
	})
	return producer
}

// Produce 创建历史记录消息
func (p *HistoryProducer) Produce(ctx context.Context, sessionId string, start, end time.Time) error {
	// 构造消息体
	msg := map[string]interface{}{
		"sessionId": sessionId,
		"start":     start.Unix(),
		"end":       end.Unix(),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	// 发布持久化消息
	err = p.channel.PublishWithContext(ctx, "chat_history_huasi", "history.huasi.end",
		false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
	return err
}

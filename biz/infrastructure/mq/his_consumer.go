package mq

import (
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xh-polaris/gopkg/util/log"
	"github.com/xh-polaris/psych-digital/biz/domain"
	"github.com/xh-polaris/psych-digital/biz/domain/model/bailian"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/config"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/mapper/history"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/rpc/psych_user"
	"github.com/xh-polaris/psych-digital/biz/infrastructure/util"
	"github.com/xh-polaris/psych-idl/kitex_gen/user"
	"golang.org/x/net/context"
)

// HistoryConsumer 消费聊天记录并生成报表
type HistoryConsumer struct {
	conn   *amqp.Connection
	finish chan struct{}
	psychU psych_user.IPsychUser
}

// NewHistoryConsumer 创建一个消费者
func NewHistoryConsumer() *HistoryConsumer {
	return &HistoryConsumer{
		conn:   getConn(),
		psychU: psych_user.NewPsychUser(config.GetConfig()),
	}
}

// Consume 启动消费者
func Consume() {
	consumer := NewHistoryConsumer()
	consumer.Start()
}

// Start 开始消费
func (c *HistoryConsumer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动消息处理
	gopool.CtxGo(ctx, func() {
		c.consume(ctx)
	})
	// 处理系统信号
	gopool.CtxGo(ctx, func() {
		c.osSignalHandler(ctx)
		c.finish <- struct{}{}
	})

	<-c.finish
}

// 消费信息
func (c *HistoryConsumer) consume(ctx context.Context) {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Error("get channel error:", err)
		return
	}
	defer func() { _ = ch.Close() }()
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Error("set qos error:", err)
		return
	}
	msgs, err := ch.Consume("chat_history_huasi", "history_consumer_huasi", false, false, false, false, nil)
	if err != nil {
		log.Error("get consume error:", err)
		return
	}

	for msg := range msgs {
		if err = c.process(ctx, msg); err != nil {
			// 失败时拒绝并重试
			log.Error("处理失败，消息重新入队:", err)
			if err = msg.Nack(false, true); err != nil {
				log.Error("nack失败 ", err)
			}
		} else if err = msg.Ack(false); err != nil {
			log.Error("ack失败 ", err)
		}
	}
}

// osSignalHandler 处理os信号
func (c *HistoryConsumer) osSignalHandler(ctx context.Context) {
	log.CtxInfo(ctx, "[osSignalHandler] start")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	osSignal := <-ch
	log.CtxInfo(ctx, "[osSignalHandler] receive signal:[%v]", osSignal)
}

// process 实际消费逻辑
func (c *HistoryConsumer) process(ctx context.Context, msg amqp.Delivery) error {
	var m map[string]interface{}
	var err error

	if err = json.Unmarshal(msg.Body, &m); err != nil {
		return err
	}

	session := m["sessionId"].(string)
	start := int64(m["start"].(float64))
	end := int64(m["end"].(float64))
	unitId := m["unitId"].(string)
	userId := m["userId"].(string)
	studentId := m["studentId"].(string)

	var res *user.UserGetInfoResp

	if res, err = c.psychU.UserGetInfo(context.Background(), &user.UserGetInfoReq{
		UserId: userId,
		UnitId: &unitId,
	}); err != nil {
		return err
	}

	rs := domain.GetRedisHelper()
	histories, err := rs.Load(session)
	if err != nil {
		return err
	}

	dialogs := make([]*history.Dialog, 0, len(histories))
	for _, his := range histories {
		dia := &history.Dialog{
			Role:    his.Role,
			Content: his.Content,
		}
		dialogs = append(dialogs, dia)
	}
	form, err := util.Anypb2Any(res.Form)
	if err != nil {
		return err
	}
	his := &history.History{
		Name:      res.User.Name,
		Class:     form["class"].(string),
		StudentId: studentId,
		Dialogs:   dialogs,
		Report:    nil,
		StartTime: time.Unix(start, 0),
		EndTime:   time.Unix(end, 0),
	}

	if len(dialogs) > 0 {
		if err = parse(his); err != nil {
			return err
		}
		// 存储对话记录
		if err = c.store(ctx, his); err != nil {
			return err
		}
	}

	// 从redis中删除
	if err = rs.Remove(session); err != nil {
		return err
	}
	return nil
}

// parse 解析对话信息
func parse(his *history.History) error {
	reportApp := bailian.GetBLReportApp()
	report, err := reportApp.Call(buildMsg(his))
	if err != nil {
		log.Error("call build error:", err)
		return err
	}
	his.Report = &history.Report{
		Keywords:   report.Report.Keywords,
		Type:       report.Report.Type,
		Content:    report.Report.Content,
		Grade:      report.Report.Grade,
		Suggestion: report.Report.Suggestion,
	}
	return err
}

// buildMsg 拼接消息
func buildMsg(his *history.History) string {
	var sb strings.Builder
	for _, h := range his.Dialogs {
		sb.WriteString(h.Role)
		sb.WriteString(":")
		sb.WriteString(h.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

// store 存储对话记录
func (c *HistoryConsumer) store(ctx context.Context, his *history.History) error {
	mapper := history.GetMongoMapper()
	return mapper.Insert(ctx, his)
}

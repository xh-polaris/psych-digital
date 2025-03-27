package volc

type Event int32

const (
	// 默认事件,对于使用事件的方案，可以通过非0值来校验事件的合法性
	EventNone Event = 0

	// 上行Connection事件
	EventStartConnection  Event = 1
	EventFinishConnection Event = 2

	// 下行Connection事件
	EventConnectionStarted  Event = 50 // 成功建连
	EventConnectionFailed   Event = 51 // 建连失败（可能是无法通过权限认证）
	EventConnectionFinished Event = 52 // 连接结束

	// 上行Session事件
	EventStartSession  Event = 100
	EventFinishSession Event = 102

	// 下行Session事件
	EventSessionStarted  Event = 150
	EventSessionFinished Event = 152
	EventSessionFailed   Event = 153

	// 上行通用事件
	EventTaskRequest Event = 200

	// 下行TTS事件
	EventTTSSentenceStart Event = 350
	EventTTSSentenceEnd   Event = 351
	EventTTSResponse      Event = 352
)

var (
	eventName = map[int32]string{
		0:   "None",
		1:   "StartConnection",
		2:   "FinishConnection",
		50:  "ConnectionStarted",
		51:  "ConnectionFailed",
		52:  "ConnectionFinished",
		100: "StartSession",
		101: "CancelSession",
		102: "FinishSession",
		150: "SessionStarted",
		151: "SessionCanceled",
		152: "SessionFinished",
		153: "SessionFailed",
		200: "TaskRequest",
		350: "TTSSentenceStart",
		351: "TTSSentenceEnd",
		352: "TTSResponse",
	}
)

func (e Event) String() string {
	return eventName[int32(e)]
}

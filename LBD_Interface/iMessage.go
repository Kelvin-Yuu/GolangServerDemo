package LBD_Interface

type MSG_TYPE uint32

const (
	// 定义您的消息类型枚举值
	IMG_SEND MSG_TYPE = iota
	IMG_RECV
	AUDIO_SEND
	AUDIO_RECV
	TEXT_SEND
	TEXT_RECV
	CREATE_MEETING
	EXIT_MEETING
	JOIN_MEETING
	CLOSE_CAMERA

	CREATE_MEETING_RESPONSE = 20
	PARTNER_EXIT            = 21
	PARTNER_JOIN            = 22
	JOIN_MEETING_RESPONSE   = 23
	PARTNER_JOIN2           = 24
	RemoteHostClosedError   = 40
	OtherNetError           = 41
)

type IMessage interface {
	GetLen() uint32       // Gets the length of the message data segment(获取消息数据段长度)
	GetMsgType() MSG_TYPE // Gets the ID of the message(获取消息ID)
	GetData() []byte      // Gets the content of the message(获取消息内容)
	//GetIp() string        // Gets the raw data of the message(获取原始数据)

	SetMsgType(MSG_TYPE) // Sets the ID of the message(设计消息ID)
	SetData([]byte)      // Sets the content of the message(设计消息内容)
	//SetIp(string2 string) // Sets the length of the message data segment(设置消息数据段长度)
}

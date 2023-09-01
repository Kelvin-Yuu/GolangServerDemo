package LBD_Interface

type IRequest interface {
	// GetConnection 得到当前连接
	GetConnection() IConnection

	// GetData 得到请求的消息数据
	GetData() []byte

	// GetMsgType 得到当前请求消息 ID
	GetMsgType() MSG_TYPE

	GetMsgLen() uint32

	//// GetMsgIp 得到当前请求的消息IP字段
	//GetMsgIp() string
}

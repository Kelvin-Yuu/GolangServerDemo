package LBD_Net

import (
	iface "GolangServerDemo/LBD_Interface"
)

type Request struct {
	// 已经和客户端建立好的连接
	conn iface.IConnection

	// 客户端请求的数据
	msg iface.IMessage
}

// GetConnection 得到当前连接
func (r *Request) GetConnection() iface.IConnection {
	return r.conn
}

// GetData 得到请求的消息数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

// GetMsgType 得到当前请求消息 ID
func (r *Request) GetMsgType() iface.MSG_TYPE {
	return r.msg.GetMsgType()
}

// GetMsgIp 得到当前请求的消息IP字段
//func (r *Request) GetMsgIp() string {
//	return r.msg.GetIp()
//}

func (r *Request) GetMsgLen() uint32 {
	return r.msg.GetLen()
}

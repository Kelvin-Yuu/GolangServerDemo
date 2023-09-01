package LBD_Interface

import (
	"context"
	"net"
)

type IConnection interface {
	//启动链接 让当前的链接准备开始工作
	Start()

	//停止链接 结束当前链接的工作
	Stop()

	//获取当前链接的绑定socket conn
	GetConnection() net.Conn

	//获取当前链接模块的链接ID
	GetConnID() uint32

	//获取消息处理器
	GetMsgHandler() IMsgHandle

	//获取远程客户端的TCP状态（IP,Port)
	RemoteAddr() net.Addr
	RemoteAddrString() string

	//获取本地服务器的TCP状态（IP,Port）
	LocalAddr() net.Addr
	LocalAddrString() string

	//直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgType MSG_TYPE, data []byte) error //添加带缓冲发送消息接口

	//返回ctx，用于用户自定义的go程获取连接退出状态
	Context() context.Context
	GetWorkerID() uint32
}

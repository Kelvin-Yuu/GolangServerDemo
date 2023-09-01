package LBD_Interface

type IServer interface {
	//启动服务器方法
	Start()
	//停止服务器方法
	Stop()
	//开启业务服务方法
	Serve()
	//得到链接管理
	GetConnMgr() IConnManager
	//获取Server绑定的消息处理模块
	GetMsgHandler() IMsgHandle

	// AddRouter 给当前的服务注册一个路由方法，供客户端的连接处理使用
	AddRouter(msgType MSG_TYPE, router IRouter)

	// 获取Server绑定的数据协议封包方式
	GetPacket() IDataPack

	// 获取服务器名称
	ServerName() string
}

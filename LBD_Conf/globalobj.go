package LBD_Conf

import (
	iface "GolangServerDemo/LBD_Interface"
)

// GlobalObj 存储一切有关框架的全局参数，供其他模块使用
type GlobalObj struct {
	/**
	Server
	*/
	TcpServer iface.IServer // 当前全局的Server对象
	Host      string        // 当前服务器主机监听的IP
	TcpPort   int           // 当前服务器主机监听的端口号
	Name      string        // 当前服务器的名称

	/**
	框架
	*/
	Version          string // 当前的版本号
	MaxConn          int    // 当前服务器主机允许的最大连接数
	MaxPackageSize   uint32 // 当前框架数据包的最大值
	WorkerPoolSize   uint32 // 当前业务工作Worker池的Goroutine数量
	MaxWorkerTaskLen uint32 // 框架允许用户最多开辟多少个Worker(限定条件)
	MaxMsgChanLen    uint32 //SendBuffMsg发送消息的缓冲最大长度
	IOReadBuffSize   uint32 //每次IO最大的读取长度
}

// GlobalObject 定义一个全局的对外对象
var GlobalObject *GlobalObj

// 初始化当前的 GlobalObject
func init() {
	GlobalObject = &GlobalObj{
		Name:             "GolangServerDemo",
		Version:          "V1.0",
		TcpPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPackageSize:   4096,
		WorkerPoolSize:   10,   // worker 工作池的队列的个数
		MaxWorkerTaskLen: 1024, // 每个worker对应的消息队列的任务的最大值
		MaxMsgChanLen:    1024,
		IOReadBuffSize:   1024,
	}
	// 尝试从 conf/zinx.json 去加载用户自定义的参数
	//GlobalObject.Reload()
}

//// Reload 从 zinx.json 加载自定义的参数
//func (g *GlobalObj) Reload() {
//	data, err := ioutil.ReadFile("conf/serverConfig.json")
//	if err != nil {
//		panic(err)
//	}
//	// 将json文件数据解析到struct中
//	err = json.Unmarshal(data, &GlobalObject)
//	if err != nil {
//		panic(err)
//	}
//}

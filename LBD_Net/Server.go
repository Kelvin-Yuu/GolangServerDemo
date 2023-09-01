package LBD_Net

import (
	conf "GolangServerDemo/LBD_Conf"
	iface "GolangServerDemo/LBD_Interface"
	lbd_log "GolangServerDemo/LBD_Log"
	"GolangServerDemo/LBD_Pack"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Server struct {
	// 服务器的名称
	Name string
	// 服务器绑定的IP版本
	IPVersion string
	// 服务器监听的IP
	IP string
	// 服务器监听的端口
	Port int

	// 当前 server 的消息管理模块，用来绑定 MsgID 和对应的处理业务 API 关系
	msgHandler iface.IMsgHandle

	// 连接管理器
	ConnMgr iface.IConnManager

	// Asynchronous capture of connection closing status
	// (异步捕获链接关闭状态)
	exitChan chan struct{}

	// connection id
	cID uint32

	// Data packet encapsulation method
	// (数据报文封包方式)
	packet iface.IDataPack
}

func NewServer() iface.IServer {
	s := &Server{
		Name:       conf.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         conf.GlobalObject.Host,
		Port:       conf.GlobalObject.TcpPort,
		msgHandler: newMsgHandle(),
		ConnMgr:    NewConnManager(),
		exitChan:   nil,

		packet: LBD_Pack.Factory().NewPack(),
	}

	return s
}

// Start 启动服务器
func (s *Server) Start() {
	lbd_log.Ins().InfoF("[Zinx] Serve Name : %s, Serve Listener at IP: %s, Port: %d\n",
		s.Name, s.IP, s.Port)
	lbd_log.Ins().InfoF("[Zinx] Version : %s, MaxConn: %d, MaxPackageSize: %d\n",
		conf.GlobalObject.Version,
		conf.GlobalObject.MaxConn,
		conf.GlobalObject.MaxPackageSize)

	lbd_log.Ins().InfoF("[START] Server Listener at IP :%s, Port %d, is starting\n", s.IP, s.Port)

	// 1. 开启消息队列和worker工作池
	s.msgHandler.StartWorkerPool()

	go s.ListenTcpConn()

}

// Stop 停止服务器
func (s *Server) Stop() {
	lbd_log.Ins().InfoF("[STOP] Zinx server name %s", s.Name)
	s.ConnMgr.ClearConn()
	s.exitChan <- struct{}{}
	close(s.exitChan)
}

// Server 运行服务器
func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	// 阻塞，否则主Go退出，listenner的go将会退出
	c := make(chan os.Signal, 1)
	// 监听指定信号 ctrl+c kill信号
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("[SERVE] Zinx server , name %s, Serve Interrupt, signal = %v \n", s.Name, sig)
}

// 得到链接管理
func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

// 获取Server绑定的消息处理模块
func (s *Server) GetMsgHandler() iface.IMsgHandle {
	return s.msgHandler
}

// AddRouter 给当前的服务注册一个路由方法，供客户端的连接处理使用
func (s *Server) AddRouter(msgType iface.MSG_TYPE, router iface.IRouter) {
	s.msgHandler.AddRouter(msgType, router)
	lbd_log.Ins().InfoF("Add Router Success!")
}

func (s *Server) ServerName() string {
	return s.Name
}

// 获取Server绑定的数据协议封包方式
func (s *Server) GetPacket() iface.IDataPack {
	return s.packet
}

func (s *Server) ListenTcpConn() {
	// 2. 获取TCP Addr
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		lbd_log.Ins().ErrorF("[START] resolve tcp addr error: %v\n", err)
		return
	}

	// 3. 监听服务器地址
	listener, err := net.ListenTCP(s.IPVersion, addr)
	if err != nil {
		lbd_log.Ins().ErrorF("[START] Listen %v, err: %v\n", s.IPVersion, err)
		return
	}
	lbd_log.Ins().InfoF("[START] Start server %v success, Listening...", s.Name)

	// 4. 阻塞等待客户端连接，处理客户端连接业务
	// 3. 启动服务端业务
	go func() {
		for {
			// 3.1 设置服务器最大连接控制，如果超过最大连接，则等待
			// TODO 高并发限流策略
			if s.ConnMgr.Len() >= conf.GlobalObject.MaxConn {
				lbd_log.Ins().InfoF("Exceeded the maxConnNum:%d, Wait:%d", conf.GlobalObject.MaxConn, AcceptDelay.duration)
				AcceptDelay.Delay()
				continue
			}
			// 3.2 阻塞等待客户端建立连接请求
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					lbd_log.Ins().ErrorF("Listener closed")
					return
				}
				lbd_log.Ins().ErrorF("Accept err: %v", err)
				AcceptDelay.Delay()
				continue
			}

			AcceptDelay.Reset()

			// 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			newCid := atomic.AddUint32(&s.cID, 1)
			dealConn := newServerConn(s, conn, newCid)

			go s.StartConn(dealConn)
		}
	}()
	select {
	case <-s.exitChan:
		err := listener.Close()
		if err != nil {
			lbd_log.Ins().ErrorF("listener close err: %v", err)
		}
	}
}

func (s *Server) StartConn(conn iface.IConnection) {
	// HeartBeat check
	//if s.hc != nil {
	//	// Clone一个心跳检测
	//	heartbeatChecker := s.hc.Clone()
	//
	//	// 绑定conn
	//	heartbeatChecker.BindConn(conn)
	//}
	// 启动conn业务
	conn.Start()
}

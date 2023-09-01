package LBD_Net

import (
	conf "GolangServerDemo/LBD_Conf"
	iface "GolangServerDemo/LBD_Interface"
	lbd_log "GolangServerDemo/LBD_Log"
	"GolangServerDemo/LBD_Pack"
	"context"
	"encoding/hex"
	"errors"
	"net"
	"sync"
	"time"
)

type Connection struct {
	// 当前 conn 属于哪个 Server
	TcpServer iface.IServer

	// 当前连接的 socket TCP 套接字
	conn net.Conn
	// 连接ID
	connID uint32
	// 当前的连接状态
	isClosed bool

	//负责处理该链接的workId
	workerID uint32

	// 通知当前连接停止的 channel(由 Reader 告知 Writer 退出)
	ExitChan chan bool

	//有缓冲管道，用于读写Goroutine之间的消息通信
	msgBuffChan chan []byte

	// 消息的管理 MsgID 和对应处理业务 API 关系
	msgHandler iface.IMsgHandle

	//当前链接是属于哪个Connection Manager的
	connManager iface.IConnManager

	// 链接名称，默认与创建链接的Server/Client的Name一致
	name string

	// 当前链接的本地地址
	localAddr string

	// 当前链接的远程地址
	remoteAddr string

	// Data packet packaging method
	// (数据报文封包方式)
	packet iface.IDataPack

	//告知该链接已经退出/停止的ctx
	ctx    context.Context
	cancel context.CancelFunc

	// Lock for user message reception and transmission
	// (用户收发消息的Lock)
	msgLock sync.RWMutex
}

func newServerConn(server iface.IServer, conn net.Conn, connID uint32) *Connection {
	c := &Connection{
		conn:        conn,
		connID:      connID,
		isClosed:    false,
		msgBuffChan: nil,
		name:        server.ServerName(),
		localAddr:   conn.LocalAddr().String(),
		remoteAddr:  conn.RemoteAddr().String(),
	}

	c.msgHandler = server.GetMsgHandler()

	//将当前的Conn与Server的ConnManager绑定
	c.connManager = server.GetConnMgr()
	c.packet = server.GetPacket()
	server.GetConnMgr().Add(c)

	return c
}

// 启动链接 让当前的链接准备开始工作
func (c *Connection) Start() {
	defer func() {
		if err := recover(); err != nil {
			lbd_log.Ins().ErrorF("Connection Start() error: %v", err)
		}
	}()
	c.ctx, c.cancel = context.WithCancel(context.Background())

	//占用workerID
	c.workerID = useWorker(c)

	//启动当前连接的读数据业务
	go c.StartReader()

	select {
	case <-c.ctx.Done():
		c.finalizer()

		//归还workerID
		freeWorker(c)
		return
	}
}

// 停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	c.cancel()
}

// 获取当前链接的绑定socket conn
func (c *Connection) GetConnection() net.Conn {
	return c.conn
}

// 获取当前链接模块的链接ID
func (c *Connection) GetConnID() uint32 {
	return c.connID
}

// 获取消息处理器
func (c *Connection) GetMsgHandler() iface.IMsgHandle {
	return c.msgHandler
}

// 获取远程客户端的TCP状态（IP,Port)
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
func (c *Connection) RemoteAddrString() string {
	return c.remoteAddr
}

// 获取本地服务器的TCP状态（IP,Port）
func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
func (c *Connection) LocalAddrString() string {
	return c.localAddr
}

// 直接将Message数据发送给远程的TCP客户端(有缓冲)
func (c *Connection) SendBuffMsg(msgType iface.MSG_TYPE, data []byte) error {
	c.msgLock.RLock()
	defer c.msgLock.RUnlock()

	if c.isClosed == true {
		return errors.New("connection closed when send buff msg")
	}
	if c.msgBuffChan == nil {
		c.msgBuffChan = make(chan []byte, conf.GlobalObject.MaxWorkerTaskLen)
		// 开启用于写回客户端数据流程的Goroutine
		// 此方法只读取MsgBuffChan中的数据没调用SendBuffMsg可以分配内存和启用协程
		go c.StartWriter()
	}
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	msg, err := c.packet.Pack(LBD_Pack.NewMsgPackage(msgType, data))
	if err != nil {
		lbd_log.Ins().ErrorF("Pack error: " + err.Error())
		return errors.New("Pack error msg: " + err.Error())
	}

	// send timeout
	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case c.msgBuffChan <- msg:
		return nil
	}

}

// 返回ctx，用于用户自定义的go程获取连接退出状态
func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) GetWorkerID() uint32 {
	return c.workerID
}

// 链接的读业务方法
func (c *Connection) StartReader() {
	lbd_log.Ins().InfoF("[Reader Goroutine is running]")
	defer lbd_log.Ins().InfoF("%s [conn Reader exit!]", c.RemoteAddr().String())
	defer c.Stop()
	defer func() {
		if err := recover(); err != nil {
			lbd_log.Ins().ErrorF("connID=%d, panic err=%v", c.GetConnID(), err)
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			buffer := make([]byte, conf.GlobalObject.IOReadBuffSize)

			// 从conn的IO中读取数据到内存缓存buffer中
			n, err := c.conn.Read(buffer)
			if err != nil {
				lbd_log.Ins().ErrorF("read msg [read datalen=%d], error = %s", n, err)
				return
			}
			lbd_log.Ins().DebugF("read buffer %s \n", hex.EncodeToString(buffer[0:n]))
			msg, err := c.packet.Unpack(buffer[0:n])
			if err != nil {
				lbd_log.Ins().ErrorF("unpack msg [read datalen=%d], error = %s", n, err)
				return
			}

			// 得到当前conn数据的Request请求
			req := Request{
				conn: c,
				msg:  msg,
			}

			if conf.GlobalObject.WorkerPoolSize > 0 {
				// 已经开启工作池，将消息发送给工作池
				c.msgHandler.SendMsgToTaskQueue(&req)
			} else {
				// 从路由中，找到注册绑定的connection对应的router调用
				// 根据绑定好的MsgID找到处理对应API业务 执行
				go c.msgHandler.DoMsgHandler(&req)
			}
		}
	}
}

// 写消息的Goroutine，专门发送给Client消息的模块
func (c *Connection) StartWriter() {
	lbd_log.Ins().InfoF("Writer Goroutine is running")
	defer lbd_log.Ins().InfoF("%s [conn Writer exit!]", c.RemoteAddr().String())

	// 不断阻塞等待channel消息
	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.conn.Write(data); err != nil {
					lbd_log.Ins().ErrorF("Send Buff Data error:, %s Conn Writer exit", err)
					return
				}
			} else {
				lbd_log.Ins().InfoF("msgBuffChan is Closed")
				break
			}
		case <-c.ctx.Done():
			//代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

func (c *Connection) finalizer() {
	// 如果用户注册了该链接， 那么这里直接调用关闭回调业务

	c.msgLock.Lock()
	defer c.msgLock.Unlock()

	if c.isClosed == true {
		return
	}

	_ = c.conn.Close()

	if c.connManager != nil {
		c.connManager.Remove(c)
	}

	if c.msgBuffChan != nil {
		close(c.msgBuffChan)
	}

	c.isClosed = true

	lbd_log.Ins().InfoF("Conn Stop()...ConnID = %d", c.connID)

}

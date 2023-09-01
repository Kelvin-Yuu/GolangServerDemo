package LBD_Net

import (
	conf "GolangServerDemo/LBD_Conf"
	iface "GolangServerDemo/LBD_Interface"
	lbd_log "GolangServerDemo/LBD_Log"
	"strconv"
	"sync"
)

// MsgHandle 消息处理模块的实现
type MsgHandle struct {
	// 存放每个MsgID对应的处理方法
	Apis map[iface.MSG_TYPE]iface.IRouter
	// 负责 Worker 取任务的消息队列
	TaskQueue []chan iface.IRequest
	// 业务工作 Worker 池的 worker 数量
	WorkerPoolSize uint32
	// 空闲worker集合，用于zconf.WorkerModeBind
	freeWorkers  map[uint32]struct{}
	freeWorkerMu sync.Mutex
}

func newMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[iface.MSG_TYPE]iface.IRouter),
		WorkerPoolSize: conf.GlobalObject.WorkerPoolSize, // 从全局配置中获取
		TaskQueue:      make([]chan iface.IRequest, conf.GlobalObject.WorkerPoolSize),
	}
}

// DoMsgHandler 调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request iface.IRequest) {
	// 1 从request中找到msgID
	handler, ok := mh.Apis[request.GetMsgType()]
	if !ok {
		lbd_log.Ins().ErrorF("api msgID = %d is NOT FOUND! Need Register!", request.GetMsgType())
	}
	// 2 根据MsgID调度router对应的业务
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// AddRouter 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgType iface.MSG_TYPE, router iface.IRouter) {
	if _, ok := mh.Apis[msgType]; ok {
		// id 已经注册
		panic("repeat api, msg ID = " + strconv.Itoa(int(msgType)))
	}
	mh.Apis[msgType] = router
	lbd_log.Ins().InfoF("Add api MsgType = %d success!", msgType)
}

// StartWorkerPool 启动 Worker 工作池
func (mh *MsgHandle) StartWorkerPool() {
	//根据workerPoolSize 分别开启Worker，每个Work用一个go来承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个Worker被启动

		//1 当前的worker对应的channel消息队列 开辟空间 第i个Worker，就用第i个TaskQueue
		mh.TaskQueue[i] = make(chan iface.IRequest, conf.GlobalObject.MaxWorkerTaskLen)
		//2 启动当前的Worker，阻塞等待消息从channel传递过来
		go mh.startOneWorker(i, mh.TaskQueue[i])
	}
}

// SendMsgToTaskQueue 将消息交给 TaskQueue，由 worker 进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request iface.IRequest) {
	// 1 将消息分配给不同的 worker
	// 根据客户端建立的 ConnID 来分配
	workerId := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	lbd_log.Ins().DebugF("Add ConnID = %d request MsgId = %d to WorkerID = %d", request.GetConnection().GetConnID(), request.GetMsgType(), workerId)
	// 2 将消息发送给对应的 worker 的 TaskQueue
	mh.TaskQueue[workerId] <- request
}

// StartOneWorker 启动一个 Worker 工作流程
func (mh *MsgHandle) startOneWorker(workerID int, taskQueue chan iface.IRequest) {
	lbd_log.Ins().DebugF("Worker ID = %d is started ...", workerID)
	// 不断的阻塞等待对应的消息队列的消息
	for {
		select {
		// 如果有消息过来，出列的就是一个客户端的request，执行当前request绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 占用workerID
func useWorker(conn iface.IConnection) uint32 {
	mh, _ := conn.GetMsgHandler().(*MsgHandle)
	if mh == nil {
		lbd_log.Ins().ErrorF("useWorker failed, mh is nil")
		return 0
	}
	mh.freeWorkerMu.Lock()
	defer mh.freeWorkerMu.Unlock()

	for k := range mh.freeWorkers {
		delete(mh.freeWorkers, k)
		return k
	}

	// 根据ConnID来分配当前的连接应该由哪个worker负责处理
	// 轮询的平均分配法则
	// 得到需要处理此条连接的workerID
	return conn.GetConnID() % mh.WorkerPoolSize
}

func freeWorker(conn iface.IConnection) {
	mh, _ := conn.GetMsgHandler().(*MsgHandle)
	if mh == nil {
		lbd_log.Ins().ErrorF("useWorker failed, mh is nil")
		return
	}

	mh.freeWorkerMu.Lock()
	defer mh.freeWorkerMu.Unlock()

	mh.freeWorkers[conn.GetWorkerID()] = struct{}{}

}

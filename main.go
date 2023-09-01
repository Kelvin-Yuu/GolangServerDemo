package main

import (
	iface "GolangServerDemo/LBD_Interface"
	"GolangServerDemo/LBD_Log"
	"GolangServerDemo/LBD_Net"
)

type TestRouter struct {
	LBD_Net.BaseRouter
}

func (tr *TestRouter) Handle(request iface.IRequest) {
	LBD_Log.Ins().DebugF("###### 读取到来自Qt的消息 ######\n")
	LBD_Log.Ins().DebugF("MsgType: %d\n", request.GetMsgType())
	LBD_Log.Ins().DebugF("MsgLen: %v\n", request.GetMsgLen())
	LBD_Log.Ins().DebugF("MsgData: %s\n", string(request.GetData()))
	//LBD_Log.Ins().DebugF("MsgIp: %v\n", request.GetMsgIp())
	LBD_Log.Ins().DebugF("###### 准备向QT返回消息 ######\n")

	err := request.GetConnection().SendBuffMsg(iface.CREATE_MEETING_RESPONSE, []byte("Connect successfully"))
	if err != nil {
		LBD_Log.Ins().ErrorF("Send To Qt err: %s\n", err.Error())
	}

}

func main() {
	s := LBD_Net.NewServer()
	// 收到Client TypeID=1的消息响应
	s.AddRouter(1, &TestRouter{})

	s.Serve()
}

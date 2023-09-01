package LBD_Pack

import iface "GolangServerDemo/LBD_Interface"

type MSG struct {
	MsgType iface.MSG_TYPE `json:"msg_type"`
	Data    []byte         `json:"data"`
	Len     uint32         `json:"len"`
	//IP      []byte         `json:"ip"`
}

func NewMsgPackage(MsgType iface.MSG_TYPE, Data []byte) *MSG {
	return &MSG{
		MsgType: MsgType,
		Data:    Data,
		Len:     uint32(len(Data)),
	}
}

func (msg *MSG) GetLen() uint32 {
	return msg.Len
}

func (msg *MSG) GetMsgType() iface.MSG_TYPE {
	return msg.MsgType
}

func (msg *MSG) GetData() []byte {
	return msg.Data
}

//func (msg *MSG) GetIp() string {
//	return string(msg.IP)
//}

func (msg *MSG) SetData(Data []byte) {
	msg.Data = Data
	msg.Len = uint32(len(Data))
}

func (msg *MSG) SetMsgType(MsgType iface.MSG_TYPE) {
	msg.MsgType = MsgType
}

//func (msg *MSG) SetIp(Ip string) {
//	msg.IP = []byte(Ip)
//}

package LBD_Pack

import (
	conf "GolangServerDemo/LBD_Conf"
	iface "GolangServerDemo/LBD_Interface"
	"GolangServerDemo/LBD_Log"
	"bytes"
	"encoding/binary"
	"errors"
)

type DataPack struct{}

// 封包拆包实例初始化方法
func NewDataPack() iface.IDataPack {
	return &DataPack{}
}

// 封包方法，压缩数据
func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	dataBuff := bytes.NewBuffer([]byte{})

	// Write the message ID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgType()); err != nil {
		return nil, errors.New("Pack MsgType err:" + err.Error())
	}

	// Write the data length
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetLen()); err != nil {
		return nil, errors.New("Pack Len err:" + err.Error())
	}

	// Write the data
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, errors.New("Pack Data err:" + err.Error())
	}

	return dataBuff.Bytes(), nil
}

// 拆包方法，解压数据
func (dp *DataPack) Unpack(binaryData []byte) (iface.IMessage, error) {
	dataBuff := bytes.NewReader(binaryData)

	msg := &MSG{}

	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.MsgType); err != nil {
		return nil, errors.New("Unpack Data err:" + err.Error())
	}
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Len); err != nil {
		return nil, errors.New("Unpack Len err:" + err.Error())
	}
	LBD_Log.Ins().DebugF("MsgType: %d, MsgLen: %d", msg.MsgType, msg.Len)
	msg.Data = make([]byte, msg.Len)
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Data); err != nil {
		return nil, errors.New("Unpack Data err:" + err.Error())
	}
	// 判断dataLen的长度是否超出我们允许的最大包长度
	if conf.GlobalObject.MaxPackageSize > 0 && msg.GetLen() > conf.GlobalObject.MaxPackageSize {
		return nil, errors.New("Too large msg data received! ")
	}

	return msg, nil

}

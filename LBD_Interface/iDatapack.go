package LBD_Interface

type IDataPack interface {
	//// GetHeadLen 获取包的头的长度
	//GetHeadLen() uint32

	// Pack 封包
	Pack(msg IMessage) ([]byte, error)

	// Unpack 拆包
	Unpack([]byte) (IMessage, error)
}

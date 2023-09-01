package LBD_Pack

import (
	iface "GolangServerDemo/LBD_Interface"
	"sync"
)

/*
生成不同封包解包方法，【单例】
*/

var pack_once sync.Once

type pack_factory struct{}

var factoryInstance *pack_factory

func Factory() *pack_factory {
	pack_once.Do(func() {
		factoryInstance = new(pack_factory)
	})
	return factoryInstance
}

// 创建一个具体的拆包解包对象
func (f *pack_factory) NewPack() iface.IDataPack {
	var dataPack iface.IDataPack

	dataPack = NewDataPack()

	return dataPack
}

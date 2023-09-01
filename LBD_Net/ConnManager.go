package LBD_Net

import (
	iface "GolangServerDemo/LBD_Interface"
	lbd_log "GolangServerDemo/LBD_Log"
	"errors"
	"sync"
)

type ConnManager struct {
	//管理连接的集合
	connections map[uint32]iface.IConnection

	connLock sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]iface.IConnection),
	}
}

// 添加链接
func (connMgr *ConnManager) Add(conn iface.IConnection) {
	// 保护共享资源 map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 将 conn 加入到 ConnManager
	connMgr.connections[conn.GetConnID()] = conn
	lbd_log.Ins().InfoF("connection add to ConnManager successfully: conn num = %d", connMgr.Len())
}

// 删除链接
func (connMgr *ConnManager) Remove(conn iface.IConnection) {
	// 保护共享资源 map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 删除连接信息
	delete(connMgr.connections, conn.GetConnID())
	lbd_log.Ins().InfoF("connection Remove ConnID=%d successfully: conn num = %d", conn.GetConnID(), connMgr.Len())
}

// 根据connID获取链接
func (connMgr *ConnManager) Get(connID uint32) (iface.IConnection, error) {
	// 保护共享资源 map，加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()
	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not FOUND")
	}
}

// 得到当前链接总数
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

// ClearConn 清除所有连接
func (connMgr *ConnManager) ClearConn() {
	// 保护共享资源 map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 删除 conn 并停止 conn 的工作
	for connID, conn := range connMgr.connections {
		// 停止
		conn.Stop()
		// 删除
		delete(connMgr.connections, connID)
	}
	lbd_log.Ins().InfoF("Clear All Connections successfully: conn num = %d", connMgr.Len())
}

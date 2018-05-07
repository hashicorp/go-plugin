package plugin

import (
	"errors"
	"net/rpc"
	"sync"
)

// RPCConnetion acts as a werapper for a rpc.Client allowing it to be replaced safely
type RPCConnection struct {
	Name string
	lock sync.RWMutex //only lock when changing rpc
	rpc  *rpc.Client
}

// Call does a read lock and then calls the underlaying rpc Call method
func (c *RPCConnection) Call(serviceMethod string, args interface{}, resp interface{}) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.rpc.Call(serviceMethod, args, resp)
}

// setRPC will replace the currently active rpc instance when it can perform a write lock
func (c *RPCConnection) setRPC(rpc *rpc.Client) error {
	if rpc == nil {
		return errors.New("Nil RPC Client Provided")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.rpc = rpc
	return nil
}

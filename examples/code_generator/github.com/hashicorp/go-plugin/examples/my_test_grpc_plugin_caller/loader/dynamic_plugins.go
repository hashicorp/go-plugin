package loader

import (
	"sync"
	"time"

	"github.com/hashicorp/go-plugin"
)

type ProcessBase struct {
	Client   *plugin.Client
	Protocol plugin.ClientProtocol
	Entry    interface{}
	LastRun  int64
}

func (p *ProcessBase) Close() {
	if p.Protocol != nil {
		p.Protocol.Close()
	}
	p.Entry = nil
	p.Protocol = nil
	if p.Client != nil {
		p.Client.Kill()
		p.Client = nil
	}
}

type DynamicPluginsBase struct {
	Plugins sync.Map
	Lock    sync.Mutex
}

func NewDynamicPlugins() *DynamicPluginsBase {
	out := &DynamicPluginsBase{Plugins: sync.Map{}}
	go out.autoUnLoad()
	return out
}

const defaultUnloadSeconds = 20 * 60

var pluginUnloadSeconds int64 = defaultUnloadSeconds

func SetPluginUnloadSeconds(seconds int64) {
	pluginUnloadSeconds = seconds
}

func (d *DynamicPluginsBase) autoUnLoad() {
	tick := time.NewTicker(time.Duration(pluginUnloadSeconds) * time.Second)
	for {
		select {
		case <-tick.C:
			toDelete := make(map[string]interface{})
			d.Plugins.Range(func(key, value interface{}) bool {
				p, ok := value.(*ProcessBase)
				if !ok {
					toDelete[key.(string)] = nil
					return true
				}
				now := time.Now().Unix()
				if pluginUnloadSeconds > 0 && now-p.LastRun > pluginUnloadSeconds {
					toDelete[key.(string)] = p
					return true
				}
				return true
			})
			if len(toDelete) > 0 {
				d.Lock.Lock()
				for k, v := range toDelete {
					p, ok := v.(*ProcessBase)
					if ok {
						p.Close()
					}
					d.Plugins.Delete(k)
					// log.Println("remove:", k)
				}
				d.Lock.Unlock()
			}
		}
	}
}

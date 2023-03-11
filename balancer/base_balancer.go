package balancer

import (
	"sync"
)

// BaseBalancer hosts 表示的需要代理的主机
type BaseBalancer struct {
	sync.RWMutex
	hosts []string
}

// Add new host to the balancer
// 实现了 balencer 接口，作为 base 给各个负载均衡器提供 balencer 接口
func (b *BaseBalancer) Add(host string) {
	b.Lock()
	defer b.Unlock()
	for _, h := range b.hosts {
		if h == host {
			return
		}
	}
	b.hosts = append(b.hosts, host)
}

// Remove new host from the balancer
func (b *BaseBalancer) Remove(host string) {
	b.Lock()
	defer b.Unlock()
	for i, h := range b.hosts {
		if h == host {
			b.hosts = append(b.hosts[:i], b.hosts[i+1:]...)
			return
		}
	}
}

// Balance selects a suitable host according
func (b *BaseBalancer) Balance(key string) (string, error) {
	return "", nil
}

// Inc .
func (b *BaseBalancer) Inc(_ string) {}

// Done .
func (b *BaseBalancer) Done(_ string) {}

// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package balancer

import (
	"errors"
)

var (
	NoHostError                = errors.New("no host")
	AlgorithmNotSupportedError = errors.New("algorithm not supported")
)

// Balancer interface is the load balancer for the reverse proxy
// 实现 工厂模式
type Balancer interface {
	// Add 为负载均衡器新增一个主机
	Add(string)
	Remove(string)
	Balance(string) (string, error)

	// Inc Done 表示对代理主机的连接数+1或者-1操作
	Inc(string)
	Done(string)
}

// Factory is the factory that generates Balancer,
// and the factory design pattern is used here
type Factory func([]string) Balancer

var factories = make(map[string]Factory)

// Build generates the corresponding Balancer according to the algorithm
func Build(algorithm string, hosts []string) (Balancer, error) {
	factory, ok := factories[algorithm]
	if !ok {
		return nil, AlgorithmNotSupportedError
	}
	return factory(hosts), nil
}

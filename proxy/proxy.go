// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/zehuamama/balancer/balancer"
)

var (
	XRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	XProxy        = http.CanonicalHeaderKey("X-Proxy")
	XForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
)

var (
	ReverseProxy = "Balancer-Reverse-Proxy"
)

// HTTPProxy refers to a reverse proxy in the balancer
// 需要实现 ServeHTTP 来实现反向代理
type HTTPProxy struct {
	// 主机与反向代理的映射
	hostMap map[string]*httputil.ReverseProxy
	// 负载均衡器
	lb balancer.Balancer

	sync.RWMutex // protect alive
	// 判断健康状态
	alive map[string]bool
}

// NewHTTPProxy create  new reverse proxy with url and balancer algorithm
// targetHosts: 代理主机集合, algorithm: 使用的负载均衡算法
func NewHTTPProxy(targetHosts []string, algorithm string) (
	*HTTPProxy, error) {

	hosts := make([]string, 0)
	hostMap := make(map[string]*httputil.ReverseProxy)
	alive := make(map[string]bool)
	// 为每一个代理主机生成一个映射关系
	for _, targetHost := range targetHosts {
		// 解析 url
		url, err := url.Parse(targetHost)
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)

		originDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originDirector(req)
			// 为请求设置头部字段
			req.Header.Set(XProxy, ReverseProxy)
			req.Header.Set(XRealIP, GetIP(req))
		}
		// GetHost return a string like IP:port
		host := GetHost(url)
		// 主机默认存活
		alive[host] = true
		hostMap[host] = proxy
		hosts = append(hosts, host)
	}

	lb, err := balancer.Build(algorithm, hosts) // 根据算法构建负载均衡器
	if err != nil {
		return nil, err
	}

	return &HTTPProxy{
		hostMap: hostMap,
		lb:      lb,
		alive:   alive,
	}, nil
}

// ServeHTTP implements a proxy to the http server
// HTTPProxy 会使用负载均衡器依据客户端访问的IP将其定向到其中的一台主机中
// 若出现错误则返回 502 BadGateway
func (h *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("proxy causes panic :%s", err)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(err.(error).Error()))
		}
	}()

	// 用负载均衡器依据客户端访问的IP将其定向到其中的一台主机中
	host, err := h.lb.Balance(GetIP(r))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
		return
	}

	h.lb.Inc(host)
	defer h.lb.Done(host)
	h.hostMap[host].ServeHTTP(w, r)
}

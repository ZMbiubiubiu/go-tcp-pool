package pool

import (
	"errors"
	"fmt"
	"sync"
)

type Option func(pool *TcpConnPool)

// 采用Go语言功能选项模式

func CreateTcpConnPool(host string, port int, options ...Option) (*TcpConnPool, error) {
	if host == "" || port == 0 {
		return nil, errors.New("invalid parameters")
	}
	pool := &TcpConnPool{
		host:         host,
		port:         port,
		mu:           sync.Mutex{},
		idleConns:    make(map[string]*TcpConn),
		maxOpenCount: 10,
		maxIdleCount: 5,
		requestChan:  make(chan *connRequest, 10),
	}

	for _, option := range options {
		option(pool)
	}

	fmt.Printf("CreateTcpConnPool :pool config:%v\n", pool)

	// 将永久运行的处理函数，放在工厂函数里面
	go pool.handleConnectionRequest()

	return pool, nil
}

func WithMaxOpenCount(num int) Option {
	return func(pool *TcpConnPool) {
		pool.maxOpenCount = num
	}
}

func WithMaxIdleCount(num int) Option {
	return func(pool *TcpConnPool) {
		pool.maxIdleCount = num
	}
}

func WithMaxQueueLength(num int) Option {
	return func(pool *TcpConnPool) {
		pool.requestChan = make(chan *connRequest, num)
	}
}

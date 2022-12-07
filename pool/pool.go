package pool

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type TcpConnPool struct {
	host      string
	port      int
	mu        sync.Mutex
	idleConns map[string]*TcpConn
	numOpen   int
	// 常量
	maxOpenCount int
	maxIdleCount int

	// A queue of connection requests
	// 当调用Get获取连接，当下却没有，将请求插入到这个channel中
	requestChan chan *connRequest
}

// connRequest wraps a channel to receive a connection
// and a channel to receive an error
type connRequest struct {
	connChan chan *TcpConn
	errChan  chan error
}

type TcpConn struct {
	id       string
	pool     *TcpConnPool
	net.Conn // The underlying TCP connection
}

// Put 如果空闲连接达到上限，直接关闭
// 否则交给连接池管理
// put() attempts to return a used connection back to the pool
// It closes the connection if it can't do so
func (p *TcpConnPool) Put(c *TcpConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.idleConns) >= p.maxIdleCount {
		c.Close()
		p.numOpen--
	} else {
		p.idleConns[c.id] = c
		fmt.Printf("put one tcp connection.\n")
	}
}

// Get 从连接池获取一个 TCP 连接
func (p *TcpConnPool) Get() (*TcpConn, error) {
	p.mu.Lock()

	// 如果有空闲连接，随机返回一个
	if len(p.idleConns) > 0 {
		for _, conn := range p.idleConns {
			// 控线连接减少1个
			delete(p.idleConns, conn.id)
			p.mu.Unlock()
			return conn, nil
		}
	}

	if p.numOpen < p.maxOpenCount {
		p.numOpen++
		p.mu.Unlock()

		newTcpConn, err := p.openNewTcpConnection()
		if err != nil {
			p.mu.Lock()
			p.numOpen--
			p.mu.Unlock()
			return nil, err
		}

		return newTcpConn, nil
	}

	p.mu.Unlock()

	// 否则，我们就得等待其他连接空闲
	// tips: 一定要为指针类型，否则就算处理完了，你也不知道
	req := &connRequest{
		connChan: make(chan *TcpConn, 1),
		errChan:  make(chan error, 1),
	}

	p.requestChan <- req

	// 等待空闲连接，否则
	// 失败
	select {
	case tcpConn := <-req.connChan:
		return tcpConn, nil
	case err := <-req.errChan:
		return nil, err
	}
}

// handleConnectionRequest() listens to the request queue
// and attempts to fulfil any incoming requests
func (p *TcpConnPool) handleConnectionRequest() {
	for req := range p.requestChan {
		var (
			requestDone = false
			hasTimeout  = false

			// 最多等待3s
			timeoutChan = time.After(3 * time.Second)
		)

		for {
			if requestDone || hasTimeout {
				break
			}
			select {
			case <-timeoutChan:
				hasTimeout = true
				req.errChan <- errors.New("connection request timeout")
			default:
				p.mu.Lock()
				// First, we try to get an idle conn.
				// If fail, we try to open a new conn.
				// If both does not work, we try again in the next loop until timeout.

				numIdle := len(p.idleConns)
				if numIdle > 0 {
					for _, conn := range p.idleConns {
						delete(p.idleConns, conn.id)
						p.mu.Unlock()
						req.connChan <- conn
						requestDone = true

						break
					}
				} else if p.numOpen < p.maxOpenCount {
					p.numOpen++
					p.mu.Unlock()

					newTcpConn, err := p.openNewTcpConnection()
					if err != nil {
						p.mu.Lock()
						p.numOpen--
						p.mu.Unlock()
					} else {
						req.connChan <- newTcpConn
						requestDone = true
					}
				} else {
					p.mu.Unlock()
				}
			}
		}

		fmt.Printf("handleConnectionRequest retrive one tcp.--------------------------------\n")
	}
}

// openNewTcpConnection() creates a new TCP connection at p.host and p.port
func (p *TcpConnPool) openNewTcpConnection() (*TcpConn, error) {
	addr := fmt.Sprintf("%s:%d", p.host, p.port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &TcpConn{
		id:   fmt.Sprintf("%v", time.Now().UnixNano()),
		pool: p,
		Conn: conn,
	}, nil
}

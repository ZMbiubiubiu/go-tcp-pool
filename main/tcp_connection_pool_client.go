package main

import (
	pool "gogo/SyntaxStudy/TCP/pool"
	"log"
	"time"
)

func main() {
	tcpPool, err := pool.CreateTcpConnPool("localhost", 8080, pool.WithMaxOpenCount(8), pool.WithMaxIdleCount(8))
	if err != nil {
		log.Fatalln("create pool failed.")
	}

	for i := 0; i < 100; i++ {
		go func() {
			conn, err := tcpPool.Get()
			if err != nil {
				log.Print(err)
			}

			conn.Write([]byte("hello"))

			time.Sleep(20 * time.Millisecond)

			tcpPool.Put(conn)
		}()
	}

	select {}

}

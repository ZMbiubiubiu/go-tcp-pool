package main

import (
	"fmt"
	"log"
	"time"

	"go-tcp-pool/pool"
)

func main() {
	tcpPool, err := pool.CreateTcpConnPool(pool.WithMaxOpenCount(10), pool.WithMaxIdleCount(10))
	if err != nil {
		log.Fatalln("create pool failed.")
	}

	for i := 0; i < 100; i++ {
		go func(i int) {
			conn, err := tcpPool.Get()
			if err != nil {
				log.Print(err)
			}

			conn.Write([]byte("hello"))

			time.Sleep(1_000 * time.Millisecond)

			fmt.Printf("finish %d task.\n", i+1)

			tcpPool.Put(conn)
		}(i)
	}

	select {}

}

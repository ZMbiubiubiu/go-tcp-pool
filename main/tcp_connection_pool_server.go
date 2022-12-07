package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("listen failed, err(%+v)", err)
	}

	var count = 0
	for {
		c, err := l.Accept()
		count++
		if err != nil {
			log.Printf("accept failed, count(%d) err(%+v)", count, err)
			continue
		}

		_ = c

		fmt.Printf("create %d connection.", count)
	}

}

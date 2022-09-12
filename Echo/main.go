package main

import (
	"log"
	"net"
)

const (
	SERVER_IF   = "0.0.0.0"
	SERVER_PORT = "13337"
	TYPE        = "tcp"
)

func handleIncomingConnection(conn net.Conn) {
	oneByteBuffer := make([]byte, 1)
	for {
		// Read one byte
		n, err := conn.Read(oneByteBuffer)
		if err != nil || n != 1 {
			break
		}

		// Write it back
		n, err = conn.Write(oneByteBuffer)
		if err != nil || n != 1 {
			break
		}
	}

	// Cleanup
	conn.Close()
}

func main() {
	listen, err := net.Listen(TYPE, SERVER_IF+":"+SERVER_PORT)
	log.Printf("Listening on %s:%s\n", SERVER_IF, SERVER_PORT)

	if err != nil {
		log.Fatal(err)
	}

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		log.Println("Established Connection")
		if err != nil {
			log.Fatal(err)
		}
		go handleIncomingConnection(conn)
	}
}

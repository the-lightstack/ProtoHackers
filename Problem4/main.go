package main

import (
	"log"
	"net"
	"strings"
)

const (
	HOST = "0.0.0.0"
	PORT = "13337"
	TYPE = "udp"
)

func handleConnection(conn net.PacketConn, addr net.Addr, line []byte, db *map[string]string) {

	// Read until newline
	lineBuffer := string(line)
	if strings.Contains(lineBuffer, "=") {

		if lineBuffer[0] == '=' {
			(*db)[""] = string(lineBuffer[1:])
		} else {
			key_val := strings.SplitN(lineBuffer, "=", 2)

			// Make sure "version" may not be modified
			if key_val[0] == "version" {
				return
			}

			(*db)[key_val[0]] = key_val[1]

		}
	} else {
		// GET
		resp := lineBuffer + "="
		resp += (*db)[string(lineBuffer)]

		conn.WriteTo([]byte(resp), addr)
	}
}

func main() {
	log.Printf("Starting server: %s:%s\n", HOST, PORT)
	pc, err := net.ListenPacket(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer pc.Close()
	database := make(map[string]string)
	database["version"] = "Light DB v1.0"

	for {
		lineBuffer := make([]byte, 1000)
		n, addr, err := pc.ReadFrom(lineBuffer)
		if err != nil {
			log.Fatal("Error while reading into line buffer")
		}
		log.Printf("RCV: %s\n", lineBuffer)

		handleConnection(pc, addr, lineBuffer[:n], &database)
	}
}

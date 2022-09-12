package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
)

const (
	HOST              = "0.0.0.0"
	PORT              = "13337"
	TYPE              = "tcp"
	MalformedResponse = "{dkd}\n"
)

func IsPrime(num float64) bool {
	// Check if it has decimals, if yes it is not a prime
	hasDecimals := num != float64(int64(num))
	if hasDecimals {
		return false
	}
	if num < 2 {
		return false
	}
	n := int64(num)

	sq_root := int(math.Sqrt(float64(num)))
	for i := 2; i <= sq_root; i++ {
		if n%int64(i) == 0 {
			return false
		}
	}

	return true
}

func ReadLine(conn net.Conn, buf *[]byte) error {
	oneByteBuf := make([]byte, 1)
	for {
		n, err := conn.Read(oneByteBuf)
		if err != nil || n != 1 {
			return errors.New("eof occured")
		}
		if oneByteBuf[0] == '\n' {
			return nil
		}

		*buf = append(*buf, oneByteBuf...)
	}
}

func handleConnection(conn net.Conn) {
	// Read until newline
	for {
		lineBuffer := make([]byte, 0)
		ReadLine(conn, &lineBuffer)
		trimmedLine := strings.TrimSpace(string(lineBuffer))
		log.Printf("Got: '%s'", trimmedLine)

		fields, err := ParseJsonToFields(trimmedLine)
		log.Printf("Fields: %v, Error: %v", fields, err)
		if err != nil {
			n, err := conn.Write([]byte(MalformedResponse))
			if err != nil || n != len(MalformedResponse) {
				log.Printf("Couldn't send malformed response: %v", err)
			}
			return
		}

		jsonReq := FieldsToValidJsonRequest(fields)
		log.Println("Json Request:", jsonReq)
		if jsonReq.Malformed {
			n, err := conn.Write([]byte(MalformedResponse))
			if err != nil || n != len(MalformedResponse) {
				log.Printf("Couldn't send malformed response: %v", err)
			}
			return
		}

		var response string
		switch isPrime := IsPrime(jsonReq.Number); {
		case isPrime:
			response = `{"method":"isPrime","prime":true}`
		case !isPrime:
			response = `{"method":"isPrime","prime":false}`
		}

		resp := append([]byte(response), []byte("\n")...)
		log.Println("Response:", string(resp))
		n, err := conn.Write(resp)
		if err != nil || n != len(resp) {
			log.Println("Error:", err)
			fmt.Println("Error writing to connection")
		}
	}
}

func main() {
	log.Println("Starting server")
	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.Println("Got Connection ...")
		go handleConnection(conn)

	}

}

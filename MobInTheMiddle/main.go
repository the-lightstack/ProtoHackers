package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
)

const (
	Type             = "tcp"
	Host             = "0.0.0.0"
	ListenPort       = "13337"
	MaxMessageLength = 10000

	RemoteChatServerPort   = 16963
	RemoteChatServerDomain = "chat.protohackers.com"
	FakeBogusCoinAddress   = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
	RegexPattern           = `^7[[:alnum:]]{25,35}$`
)

func ReplaceAddress(msg string) string {
	re := regexp.MustCompile(RegexPattern)

	if len(msg) > 0 {
		msg = msg[:len(msg)-1]
	}

	// Split
	parts := strings.Split(msg, " ")
	for i, p := range parts {
		if re.MatchString(p) {
			parts[i] = FakeBogusCoinAddress
		}
	}

	return strings.Join(parts, " ") + "\n"
}

func ReadMessage(conn net.Conn, buf *[]byte) error {
	oneByteBuf := make([]byte, 1)
	for i := 0; i < MaxMessageLength; i++ {
		_, err := conn.Read(oneByteBuf)
		if err != nil {
			return err
		}

		*buf = append(*buf, oneByteBuf...)

		// Include newline into message
		if oneByteBuf[0] == '\n' {
			return nil
		}
	}
	return errors.New("message too long")
}

func handleClientToServer(clientConn net.Conn, serverConn net.Conn) {
	// Always read from clientConn, replace and then write to server Conn
	for {
		tempBuf := make([]byte, 0)
		err := ReadMessage(clientConn, &tempBuf)
		if err != nil {
			log.Println("ERROR: cant read from client")
			clientConn.Close()
			serverConn.Close()
			return
		}

		// Replace
		replacedMsg := ReplaceAddress(string(tempBuf))

		if len(replacedMsg) > 0 && replacedMsg[len(replacedMsg)-1] != '\n' {
			replacedMsg += "\n"
		}
		log.Print(replacedMsg)

		// And relay to server
		_, err = serverConn.Write([]byte(replacedMsg))
		if err != nil {
			log.Println("ERROR: cant write to server!")
			clientConn.Close()
			serverConn.Close()
			return
		}
	}
}

func handleServerToClient(clientConn net.Conn, serverConn net.Conn) {
	for {
		// Read from server
		tempBuf := make([]byte, 0)
		err := ReadMessage(serverConn, &tempBuf)
		if err != nil {
			log.Println("ERROR: cant read from server")
			clientConn.Close()
			serverConn.Close()
			return
		}

		// Replace
		replacedMsg := ReplaceAddress(string(tempBuf))

		if len(replacedMsg) > 0 && replacedMsg[len(replacedMsg)-1] != '\n' {
			replacedMsg += "\n"
		}

		log.Print(replacedMsg)
		// And relay to client
		_, err = clientConn.Write([]byte(replacedMsg))
		if err != nil {
			log.Println("ERROR: cant write to client!")
			clientConn.Close()
			serverConn.Close()
			return
		}
	}
}

func handleIncomingConnection(clientConn net.Conn) {

	// Establish connection to real chat server
	log.Printf("%s:%d", RemoteChatServerDomain, RemoteChatServerPort)
	chatServerConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", RemoteChatServerDomain, RemoteChatServerPort))
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Making sure all connections get closed
	defer clientConn.Close()
	defer chatServerConn.Close()

	go handleClientToServer(clientConn, chatServerConn)
	go handleServerToClient(clientConn, chatServerConn)

	zeroBuf := make([]byte, 0)
	for {
		_, err := chatServerConn.Read(zeroBuf)
		if err != nil {
			break
		}

		_, err = clientConn.Read(zeroBuf)
		if err != nil {
			break
		}
	}
}

func main() {

	listen, err := net.Listen(Type, Host+":"+ListenPort)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	log.Printf("Started serving on %s:%s\n", Host, ListenPort)

	for {
		conn, err := listen.Accept()
		log.Println("Got Connection")
		if err != nil {
			log.Fatal(err)
		}

		go handleIncomingConnection(conn)
	}
}

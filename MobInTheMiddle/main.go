package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
)

const (
	Type             = "tcp"
	Host             = "0.0.0.0"
	ListenPort       = "13337"
	MaxMessageLength = 10000

	RemoteChatServerPort   = 16963
	RemoteChatServerDomain = "chat.protohackers.com"
	FakeBogusCoinAddress   = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
)

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

// Writes to the server_send pipe after applying regex
// on bogus coin address and replacing it
func handleClientRead(conn net.Conn, serverSend chan string, exit chan int) {
	for {
		select {
		case n := <-exit:
			exit <- n + 1
			return
		default:
			buf := make([]byte, 0)
			err := ReadMessage(conn, &buf)
			if err != nil {
				log.Println("Error receiving from client")
				exit <- 0
				return
			}

			log.Printf("Got from Client: %s", buf)
			re := regexp.MustCompile(`(^[[:space:]]?7[[:alnum:]]{25,35})|(7[[:alnum:]]{25,35})[[:space:]]?$`)
			fake := re.ReplaceAllString(string(buf), FakeBogusCoinAddress)

			serverSend <- string(fake)
		}

	}
}

func handleClientSend(conn net.Conn, clientSend chan string, exit chan int) {
	// Reads from "clientSend" and relays it
	for {
		select {
		case n := <-exit:
			select {
			case msg := <-clientSend:
				log.Println("SENDING SPECIAL SHIT TO **CLIENT** FROM DEAD CODE!!!")

				_, err := conn.Write([]byte(msg))
				if err != nil {
					log.Println("Sending to client failed :shrug:")

				}
				exit <- n + 1
				return
			default:
				exit <- n + 1
				return

			}
		default:
			msg := <-clientSend
			log.Printf("Sending to Client: %s", string(msg))

			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Sending to client failed :shrug:")
				exit <- 0
				return
			}
		}

	}
}

func handleChatServerSend(chatServerConn net.Conn, sendToServer chan string, exit chan int) {
	for {
		select {
		case n := <-exit:
			select {
			case msg := <-sendToServer:

				log.Println("SENDING SPECIAL SHIT TO SERVER FROM DEAD CODE!!!")
				log.Printf("Sending to server: %s", string(msg))
				_, err := chatServerConn.Write([]byte(msg))
				if err != nil {
					log.Println("Sending to server failed")
				}
				exit <- n + 1
				return
			default:
				exit <- n + 1
				return
			}

		default:
			msg := <-sendToServer
			log.Printf("Sending to server: %s", string(msg))
			_, err := chatServerConn.Write([]byte(msg))
			if err != nil {
				log.Println("Sending to server failed")
				exit <- 0
				return
			}
		}
	}
}

func handleChatServerReceive(chatServerConn net.Conn, sendToUserChan chan string, exit chan int) {
	for {
		select {
		case n := <-exit:
			exit <- n + 1
			return

		default:
			buf := make([]byte, 0)
			err := ReadMessage(chatServerConn, &buf)
			if err != nil {
				log.Println("Error receiving from ChatServer")
				exit <- 0
				return
			}
			log.Printf("Got from Server: %s", buf)

			re := regexp.MustCompile(`(^[[:space:]]?7[[:alnum:]]{25,35})|(7[[:alnum:]]{25,35})[[:space:]]?$`)
			fake := re.ReplaceAllString(string(buf), FakeBogusCoinAddress)

			sendToUserChan <- string(fake)
		}
	}
}

func handleIncomingConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Establish connection to real chat server
	log.Printf("%s:%d", RemoteChatServerDomain, RemoteChatServerPort)
	chatServerConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", RemoteChatServerDomain, RemoteChatServerPort))
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer chatServerConn.Close()

	// Just have 2 chans
	sendToUserChan := make(chan string)
	sendToServerChan := make(chan string)

	// And one management chan to kill all go routines on closed conn
	exit := make(chan int)

	// Start all handlers
	go handleClientRead(clientConn, sendToServerChan, exit)
	go handleChatServerSend(chatServerConn, sendToServerChan, exit)
	go handleChatServerReceive(chatServerConn, sendToUserChan, exit)
	go handleClientSend(clientConn, sendToUserChan, exit)

	// The "master" quits when the "exit" chan contains a number that is 3 or
	// higher, meaning all handlers are dead
	for {
		select {
		case n := <-exit:
			if n > 3 {
				log.Println("========================================= quitting and closing conns")
				return
			} else {
				exit <- n
			}
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

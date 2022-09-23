package main

import (
	"errors"
	"log"
	"net"
	"strings"
	"unicode"
)

const (
	Type = "tcp"
	Host = "0.0.0.0"
	Port = "13337"

	MinUnameLength    = 1
	MaxUnameLength    = 50
	MaxMessageLength  = 1005
	WelcomeMessage    = "Welcome to this DeLightFull Chat Room! What is your name?\n"
	UserJoinedMessage = "* %s joined this chat room\n"
	UserLeavesMessage = "* %s left the chat room\n"
)

func ReadUsername(conn net.Conn, buf *[]byte) error {
	oneByteBuf := make([]byte, 1)
	for i := 0; i < MaxUnameLength; i++ {
		_, err := conn.Read(oneByteBuf)
		if err != nil {
			return err
		}

		// Newline terminates username
		if oneByteBuf[0] == '\n' {
			return nil
		}

		// Making sure character is alphanumeric
		if !(unicode.IsLetter(rune(oneByteBuf[0])) || unicode.IsDigit(rune(oneByteBuf[0]))) {
			return errors.New("username character not alphanumeric")
		}

		// If all conditions are met, copy over byte into buffer
		*buf = append(*buf, oneByteBuf[0])
	}

	// Username is too long if we haven't quit yet
	return errors.New("username too long")
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

func handleIncomingConnection(conn net.Conn, chatRoom *ChatRoom) {
	defer conn.Close()
	// Send them a welcoming Message
	conn.Write([]byte(WelcomeMessage))

	// Ask for their name
	var username []byte
	err := ReadUsername(conn, &username)
	if err != nil {
		// Send error message to user
		conn.Write([]byte(err.Error()))
		return
	}

	// Finally add User to chatRoom (after stripping away whitespace)
	// Make sure "username" is not after this, else potential vuln
	user := User{
		name:     []byte(strings.TrimSpace(string(username))),
		receiver: make(chan string),
		sender:   make(chan string),
		chatRoom: chatRoom,
		exit:     make(chan bool),
	}

	go user.StartSendHandler(conn)
	go user.StartReceiveHandler(conn)
	go user.CheckConnectionDead(conn)

	err = chatRoom.AddUser(&user)
	if err != nil {
		log.Println("Error while adding user: ", err.Error())
		return
	}

	for {
		<-user.exit
		log.Println("Leaving: ", string(user.name))
		user.exit <- true
		chatRoom.UserLeave(&user)
		return
	}

}

func main() {

	chatRoom := ChatRoom{
		amountUsers: 0,
		sendMessage: make(chan Message),
	}
	go chatRoom.HandleMessageSpreading()

	listen, err := net.Listen(Type, Host+":"+Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	log.Printf("Started serving on %s:%s\n", Host, Port)

	for {
		conn, err := listen.Accept()
		log.Println("Got Connection")
		if err != nil {
			log.Fatal(err)
		}

		go handleIncomingConnection(conn, &chatRoom)
	}
}

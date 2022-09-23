package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/exp/slices"
)

type ChatRoom struct {
	amountUsers int32
	users       []*User
	sendMessage chan Message
}

type User struct {
	name     []byte
	receiver chan string
	sender   chan string
	chatRoom *ChatRoom
	exit     chan bool
}

func (u *User) String() string {
	return string(u.name)
}

type Message struct {
	senderName string
	message    string
}

// Continously checks if connection is dead
func (user *User) CheckConnectionDead(conn net.Conn) {
	zero := make([]byte, 0)
	for {
		select {
		case <-user.exit:
			user.exit <- true
			return
		default:
			_, err := conn.Read(zero)
			if err != nil {
				user.exit <- true
				return
			}
		}

	}
}

// Puts everything received from the connection into the receiver chan
func (user *User) StartReceiveHandler(conn net.Conn) {

	var msg []byte
	for {
		select {
		case <-user.exit:
			user.exit <- true
			log.Println("Actually exiting user: ", string(user.name))
			return

		default:
			err := ReadMessage(conn, &msg)
			if err != nil {
				user.exit <- true
				log.Println("Actually exiting user: ", string(user.name))
				return
			}

			log.Print("Message: ", string(msg))
			if string(msg) == "exit\n" {
				log.Println("Trigger exit!")
				user.exit <- true
				return
			}

			user.chatRoom.sendMessage <- Message{
				senderName: string(user.name),
				message:    string(msg),
			}
			msg = nil
		}
	}

}

// Writes everything from "sender" chan to the client
func (user *User) StartSendHandler(conn net.Conn) {
	for {
		select {
		case <-user.exit:
			log.Println("Actually exiting user: ", string(user.name))
			user.exit <- true
			return
		default:
			msg := <-user.sender
			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Println("Sending exit for user: ", string(user.name))
				user.exit <- true
				return
			}
		}

	}

}

func (cr *ChatRoom) UserLeave(user *User) {
	log.Println("Calling leave on :", string(user.name))

	// Find slice index of user
	var removeUserIndex uint32
	for i, u := range cr.users {
		if string(user.name) == string(u.name) {
			removeUserIndex = uint32(i)
			break
		}
	}

	// Remove user from slice
	cr.users = slices.Delete(cr.users, int(removeUserIndex), int(removeUserIndex+1))

	// Send everyone a message that user left ()
	leaveMessage := fmt.Sprintf(UserLeavesMessage, user.name)
	for _, u := range cr.users {
		u.sender <- leaveMessage
	}
}

func (cr *ChatRoom) AddUser(user *User) error {
	for _, otherUser := range cr.users {
		if string(otherUser.name) == string(user.name) {
			user.exit <- true
			return errors.New("username already exists in chat room")
		}
	}
	log.Printf("User %s Joined. ", user.name)

	announceUserMessage := fmt.Sprintf(UserJoinedMessage, string(user.name))

	// Construct message that contains all users in room
	messageToNewUser := "* Users in Room: "

	// Announce to everyone
	for _, otherUser := range cr.users {
		if string(otherUser.name) != string(user.name) {
			otherUser.sender <- announceUserMessage
		}
		messageToNewUser += fmt.Sprintf("%s, ", otherUser.name)
	}
	messageToNewUser = strings.TrimSuffix(messageToNewUser, ", ") + "\n"

	cr.users = append(cr.users, user)
	cr.amountUsers++

	// Finally send new user msg of all users that are in the room
	user.sender <- messageToNewUser
	return nil
}

func (cr *ChatRoom) HandleMessageSpreading() {
	for {
		message := <-cr.sendMessage

		formattedMessage := fmt.Sprintf("[%s] %s", message.senderName, message.message)
		// Send to all users but sender
		for _, user := range cr.users {
			if string(user.name) != message.senderName {
				user.sender <- formattedMessage
			}
		}

	}
}

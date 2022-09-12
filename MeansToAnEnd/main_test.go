package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"testing"
)

const SERVER_PORT = "13337"
const SERVER_IF = "localhost"

type Message struct {
	Type   uint8
	Field1 uint32
	Field2 uint32
}

func SerializeMessage(msg *Message) []byte {
	buf := make([]byte, 9)
	field1 := make([]byte, 4)
	field2 := make([]byte, 4)

	buf[0] = msg.Type
	binary.BigEndian.PutUint32(field1, msg.Field1)
	binary.BigEndian.PutUint32(field2, msg.Field2)

	copy(buf[1:], field1)
	copy(buf[5:], field2)

	if !(buf[0] == 'I' || buf[0] == 'Q') {
		log.Fatal("Identifier must be either 'I' or 'Q'")
	}
	if len(buf) != 9 {
		log.Fatal("Length of serialized message not 9 bytes")
	}
	return buf
}

func TestMain(m *testing.T) {
	go main()
}

func TestSerializeMessage(t *testing.T) {
	inputs := []Message{{'I', 1000, 0xffffffff}, {'Q', 0xffffedab, 0x1337}}
	outputs := []([]byte){[]byte{73, 0, 0, 3, 232, 255, 255, 255, 255},
		[]byte{81, 255, 255, 237, 171, 0, 0, 19, 55},
	}

	for i := 0; i < len(inputs); i++ {
		out := SerializeMessage(&inputs[i])
		if !Equal(outputs[i], out) {
			t.Error("Wrong Serialization of Message")
		}
	}

}

func TestConnectivity(t *testing.T) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", SERVER_IF, SERVER_PORT))
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("%v\n", conn)

	log.Println("Testing Query on empty DB")
	response := make([]byte, 4)
	conn.Write(SerializeMessage(&Message{Type: 'Q', Field1: 1000, Field2: 2}))
	conn.Read(response)
	intResponse := binary.BigEndian.Uint32(response)

	if intResponse != 0 {
		t.Error("Query on Empty Db for client should return 0")
	}

	log.Println("Testing after one inserted value")
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 1000, Field2: 500}))
	conn.Write(SerializeMessage(&Message{Type: 'Q', Field1: 900, Field2: 1100}))

	conn.Read(response)
	intResponse = binary.BigEndian.Uint32(response)
	if intResponse != 500 {
		t.Error("Query on DB with one inserted element returned different response")
	}

	log.Println("Getting mean of two inserted Values")
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 1001, Field2: 600}))
	conn.Write(SerializeMessage(&Message{Type: 'Q', Field1: 900, Field2: 1100}))

	conn.Read(response)
	intResponse = binary.BigEndian.Uint32(response)
	if intResponse != 550 {
		t.Error("Query on DB with one inserted element returned different response")
	}

	log.Println("Decimal result, which has to be rounded down")
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 1002, Field2: 3}))
	conn.Write(SerializeMessage(&Message{Type: 'Q', Field1: 900, Field2: 1100}))

	conn.Read(response)
	intResponse = binary.BigEndian.Uint32(response)
	if intResponse != 367 {
		t.Error("Query on DB with one inserted element returned different response")
	}
	conn.Close()

	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", SERVER_IF, SERVER_PORT))
	if err != nil {
		t.Fatal(err)
		return
	}

	response = make([]byte, 4)
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 0, Field2: 2000000000}))
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 1, Field2: 2050000000}))
	conn.Write(SerializeMessage(&Message{Type: 'I', Field1: 2, Field2: 2100000000}))

	conn.Write(SerializeMessage(&Message{Type: 'Q', Field1: 0, Field2: 2}))

	ReadComplete(conn, &response)
	intResponse = binary.BigEndian.Uint32(response)
	if intResponse != 2050000000 {
		t.Fatal("Mean of three numbers failed")
	}

}

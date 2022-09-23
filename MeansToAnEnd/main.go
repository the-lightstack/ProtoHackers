package main

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"
)

const (
    Host = "0.0.0.0"
	Port = "13337"
	Type = "tcp"
)

type StockData struct {
	Timestamp int32
	Price     int32
}

type QueryMessage struct {
	MinTime int32
	MaxTime int32
}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func InsertStockData(stockPriceDb *[]StockData, msg StockData, ipString string) {
	*stockPriceDb = append(*stockPriceDb, msg)

}

func PrintAsHex(b []byte) {
	log.Println(">>", hex.EncodeToString(b))
}

func QueryStockData(stockPriceDb *[]StockData, msg QueryMessage, ipString string) int32 {
	var foundPrices []int32

	// Finding all Prices that match criteria
	for _, data := range *stockPriceDb {

		if msg.MinTime <= data.Timestamp && data.Timestamp <= msg.MaxTime {
			foundPrices = append(foundPrices, data.Price)
		}
	}

	// Getting sum of all Prices
	var sum int64
	sum = 0
	for _, v := range foundPrices {
		sum += int64(v)
	}
	// Else we would have a division by zero
	if len(foundPrices) == 0 {
		return 0
	}
	mean := int32(sum / int64(len(foundPrices)))
	log.Println("> ", mean)
	return mean
}

// n must be the length of buf
func ReadComplete(conn net.Conn, buf *[]byte) error {
	tempBuffer := make([]byte, 1)
	for i := 0; i < len(*buf); i++ {
		n, err := conn.Read(tempBuffer)
		if err != nil {
			return err
		}
		if n == 0 {
			log.Println("EOF ?!?")
		}
		(*buf)[i] = tempBuffer[0]
	}
	return nil
}

func handleIncomingConnection(conn net.Conn) {
	var stockPriceDb []StockData

	clientIdentifier := conn.RemoteAddr().String()
	for {
		// Break on EOL or something, idk
		messageBuffer := make([]byte, 9)
		// n, err := conn.Read(messageBuffer)
		err := ReadComplete(conn, &messageBuffer)

		if err != nil {
			log.Println("(Read) Returning.", err)
			return
		}

		switch messageBuffer[0] {
		case 'I':
			timestamp := int32(binary.BigEndian.Uint32(messageBuffer[1:5]))
			price := int32(binary.BigEndian.Uint32(messageBuffer[5:9]))

			stockData := StockData{Timestamp: timestamp, Price: price}
			log.Println("Insert: ", stockData)
			InsertStockData(&stockPriceDb, stockData, clientIdentifier)

		case 'Q':
			minTime := int32(binary.BigEndian.Uint32(messageBuffer[1:5]))
			maxTime := int32(binary.BigEndian.Uint32(messageBuffer[5:9]))

			queryMessage := QueryMessage{MinTime: minTime, MaxTime: maxTime}
			log.Println("Query: ", queryMessage)
			meanPrice := QueryStockData(&stockPriceDb, queryMessage, clientIdentifier)

			meanPriceAsBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(meanPriceAsBytes, uint32(meanPrice))
			PrintAsHex(meanPriceAsBytes)
			n, err := conn.Write(meanPriceAsBytes)
			if err != nil || n != 4 {
				log.Println("(Write) Returning: ", err, n)
				return
			}
		default:
		}

	}

}

func main() {
	// file, _ := os.OpenFile("/dev/null", os.O_RDWR, 0666)
	// log.SetOutput(file)
	listen, err := net.Listen(Type, Host+":"+Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	// Hoping this will grow
	log.Printf("Started serving on %s:%s\n", Host, Port)

	for {
		conn, err := listen.Accept()
		log.Println("Got Connection")
		if err != nil {
			log.Fatal(err)
		}
		go handleIncomingConnection(conn)
	}
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type IncomingData struct {
	senderIP string
	data     []byte
}

type LabelData struct {
	serialNumber string
	weight       string
	targetIP     string
	revLevel     string
	interimID    int
}

//socket listen on port 44818 for incoming data
//creates new go routine for every connection accepted
func Listen() {
	port := 44818

	listen, err := net.Listen("tcp", ":44818")
	if err != nil {
		log.Fatalf("listen on port %d failed,%s", port, err)
		os.Exit(1)
	}

	defer listen.Close()

	fmt.Println("listening on port 44818.......")

	for {
		conn, err := listen.Accept()
		if logging {
			fmt.Println("connection accepted")
		}
		if err != nil {
			log.Fatalln(err)
			continue
		}

		go handler(conn)
	}

}

//handles connections in a go routine
func handler(conn net.Conn) {

	var buf = make([]byte, 1024)
	var data []byte

	//using loop to keep connection open until last of expected data is received
	for {
		n, err := conn.Read(buf)
		if logging {
			fmt.Println("Data received")
		}
		if err != nil {
			log.Println(err)
			break
		}

		if n < 10 {
			continue
		}

		data = make([]byte, n)
		copy(data, buf[:n])

		//ignore everything that has an unexpected value in 1st byte
		if string(data[0]) != "#" && data[0] != 0x04 && data[0] != 0x65 && data[0] != 0x6f {
			if logging {
				fmt.Println("invalid data received")
			}

			break //break out of the for loop to close connection
		}

		//# is used as a prefix when sending data to this program from tcp client.
		if string(data[0]) == "#" {
			if logging {
				fmt.Println("data received from tcp client")
			}
			parseIncomingData(IncomingData{senderIP: "local", data: data})
			break
		}

		// plc sending ListServices request with 0x04 in first byte
		// not required in all processors
		// from CIP Network Library Vol 2 section 2-4.6
		if data[0] == 0x04 {
			if logging {
				fmt.Println("plc message list services")
			}
			data[2] = 26
			resArr := []byte{0x01, 0x00, 0x00, 0x01, 0x14, 0x00, 0x01, 0x00, 0x20, 0x00}
			nameOfService := []byte{0x43, 0x6f, 0x6d, 0x6d, 0x75, 0x6e, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x00, 0x00}

			data = append(data, resArr...)
			data = append(data, nameOfService...)
			//time.Sleep(time.Millisecond * 50)
			conn.Write(data)

		}

		// plc is requesting register session with 0x65 in first byte
		// from CIP Network Library Vol 2 section 2-4.4
		// generate a 4 byte session handle and return to requesting device in words 4-7.
		// the rest of the return is echoing back what is received.
		if data[0] == 0x65 {
			if logging {
				fmt.Println("plc message session registered")
			}
			sessionHandle := int32To4ByteLittleEndian(getRandomInt(63000))

			data[4] = sessionHandle[0]
			data[5] = sessionHandle[1]
			data[6] = sessionHandle[2]
			data[7] = sessionHandle[3]
			time.Sleep(time.Millisecond * 50)
			conn.Write(data)
			//fmt.Println(data)
		}

		// plc sending SendRRData Request with 0x6f in first byte
		// from CIP Network Library Vol 2 section 2-4.7
		// acknowledge response is echoing the first 44 words of the request, changing word 2 to value of 0x14
		// values for words 38 - 43 must be changed before sending the response
		// word 48 is the length of the complete write request
		// word 50 begins the write request Publication 1756-PM020A-EN-P Logix5000 DAta Access page 22
		//
		// received data is sent to parseIncomingData() in go routine and the break
		// is used to exit the for loop and close the connection
		if data[0] == 0x6f {

			if logging {
				fmt.Println("plc message rr header")
			}

			resp := data[:44]
			resp[2] = 0x14
			resp[3] = 0x00
			resp[38] = 0x04
			resp[39] = 0x00
			resp[40] = 0xcd
			resp[41] = 0x00
			resp[42] = 0x00
			resp[43] = 0x00
			time.Sleep(time.Millisecond * 50)
			conn.Write(resp)

			writeReqLength := int(data[48])

			dataResult := IncomingData{"ip address", data[50 : writeReqLength+50]}

			go parseIncomingData(dataResult)

			break
		}
	}

	conn.Close()

}

//parses data in accordance with Publication 1756-PM020A-EN-P Logix5000 DAta Access for ethernet/ip communication
//parses comma delimited string for tcp communications
func parseIncomingData(dataToParse IncomingData) {

	if logging {
		fmt.Println("parsing data")
	}
	var labelData = LabelData{}

	//local keyword is used to indicate tcp communication
	//parse incoming data as string
	if dataToParse.senderIP == "local" {

		s := string(dataToParse.data[1:])
		str := strings.TrimSpace(s)

		strArr := strings.Split(str, ",")
		if len(strArr) != 5 {
			fmt.Println("invalid format. Required Format:\nserial number, weight, printer ip address, rev level, interim flag")
			return
		}

		//assign results to labelData struct for use when printing label
		labelData.serialNumber = strArr[0]
		labelData.weight = strArr[1]
		labelData.targetIP = strArr[2]
		labelData.revLevel = strings.TrimSpace(strArr[3])
		v, err := strconv.Atoi(strings.Trim(strArr[4], "\x00"))
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		labelData.interimID = v

		printLabel(labelData)
		return
	}

	//parse data received by ethernet/ip
	dataLen := dataToParse.data[1]
	dataTypePos := (dataLen * 2) + 2
	dataType := int(dataToParse.data[dataTypePos])

	//dataType of 0xa0 indicates string sent from plc
	//two string types will be accepted normal string standard to RLLogix data types
	//and a user defined string with length of 200 characters. Any other string types
	//received will be ignored
	if dataType == 0xa0 {
		extDataType := dataToParse.data[dataTypePos : dataTypePos+4]

		if extDataType[0] == 0xa0 && extDataType[1] == 0x02 && extDataType[2] == 0xce && extDataType[3] == 0x0f ||
			extDataType[0] == 0xa0 && extDataType[1] == 0x02 && extDataType[2] == 0xdb && extDataType[3] == 0x63 {

			stringLenPointer := int(dataTypePos) + 6
			sl := []byte{dataToParse.data[stringLenPointer], dataToParse.data[stringLenPointer+1], dataToParse.data[stringLenPointer+2], dataToParse.data[stringLenPointer+3]}
			stringLen := int((sl[3] << 24) + (sl[2] << 16) + (sl[1] << 8) + sl[0])

			s := string(dataToParse.data[stringLenPointer+4 : stringLenPointer+4+stringLen])
			str := strings.Trim(s, "\x00")
			strArr := strings.Split(str, ",")

			//labelData struct is populated with the results
			labelData.serialNumber = strArr[0]
			labelData.weight = strArr[1]
			labelData.targetIP = strArr[2]
			labelData.revLevel = strings.TrimSpace(strArr[3])
			v, err := strconv.Atoi(string(strArr[4]))
			if err != nil {
				fmt.Println("ERROR:", err)
			}

			labelData.interimID = v

			printLabel(labelData)

		}
	} else {

		//everything in this else clause handles the data the plc sent as an array of INTs
		//or an array of DINTs. If any other data type is sent, it will be ignored (expect
		//STRING or user defined 200 character string as mentioned above)
		dataSize := getDataSize(dataType)

		dataLen := int(dataToParse.data[dataTypePos+2])
		dataValues := dataToParse.data[dataTypePos+4 : int(dataTypePos)+4+(dataLen*dataSize)]

		var parsedDataValues []int

		for i := 0; i < len(dataValues); i += dataSize {

			switch dataType {
			case 0xc2:
				parsedDataValues = append(parsedDataValues, int(dataValues[i]))
			case 0xc3:
				parsedDataValues = append(parsedDataValues, int(dataValues[i])+int(dataValues[i+1])<<8)
			case 0xc4:
				parsedDataValues = append(parsedDataValues, int(dataValues[i])+int(dataValues[i+1])<<8+int(dataValues[i+2])<<16+int(dataValues[i+3])<<24)
			default:
				fmt.Println("switch default")
			}
		}

		var str string
		for _, i := range parsedDataValues[:12] {
			str += string(i)
		}

		labelData.serialNumber = str

		//weight is formatted to xxx.xx
		ww := float64(parsedDataValues[12])
		wd := float64(parsedDataValues[13]) * .01

		//labelData struct is populated with results
		labelData.weight = fmt.Sprintf("%.2f", ww+wd)
		labelData.targetIP = fmt.Sprintf("%d.%d.%d.%d", parsedDataValues[14], parsedDataValues[15], parsedDataValues[16], parsedDataValues[17])
		labelData.revLevel = fmt.Sprintf("%02d", parsedDataValues[18])
		labelData.interimID = int(parsedDataValues[19])

		printLabel(labelData)

	}
}

func getDataSize(dataType int) int {

	switch dataType {
	case 0xc1, 0xc2:
		return 1
	case 0xc3:
		return 2
	case 0xc4, 0xca, 0xd3:
		return 4
	case 0xc5:
		return 8
	}
	return 0
}

func getRandomInt(max int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(max)
}

func int32To4ByteLittleEndian(i int) []byte {
	var b []byte
	b = append(b, byte(i&0x000000ff))
	b = append(b, byte(i&0x0000ff00)>>8)
	b = append(b, byte(i&0x00ff0000)>>16)
	b = append(b, byte(i&0x7f000000)>>24)

	return b
}

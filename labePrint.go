package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

const PORT = 9100

//
func printLabel(data LabelData) {

	if logging {
		fmt.Println("opeing print code file")
	}
	printCode, err := getPrintCode(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	if logging {
		fmt.Println("replacing variables")
	}
	replaceVariables(&printCode, data)

	if logging {
		fmt.Println("Print Code:")
		fmt.Println(printCode)
	}

	sendToPrinter(printCode, data.targetIP)

}

func sendToPrinter(printCode string, ipAddress string) {

	if logging {
		fmt.Println("sending print code to printer at: ", ipAddress, ":", PORT)
	}

	addr := fmt.Sprintf("%s:%d", ipAddress, PORT)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)

	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write([]byte(printCode))

	if err != nil {
		fmt.Println(err)
	}

	conn.Close()

	fmt.Println("connection closed")

}

func getPrintCode(labelData LabelData) (string, error) {

	//var tempFile string
	var fileName string

	if labelData.interimID == 0 {
		tempFile := fmt.Sprintf("%s.txt", labelData.serialNumber[:2])
		fileName = path.Join(settings.FinalPrintCodeFolder, tempFile)
	} else {
		tempFile := fmt.Sprintf("%d.txt", labelData.interimID)
		fileName = path.Join(settings.InterimPrintCodeFolder, tempFile)
	}

	//fileName := path.Join(settings.FinalPrintCodeFolder, tempFile)

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}

	return string(data), err

}

func replaceVariables(pc *string, d LabelData) {

	m := time.Now().Month()
	month := fmt.Sprintf("%02d", m)
	y := time.Now().Year()
	year := fmt.Sprintf("%d", y)

	*pc = strings.Replace(*pc, "_SerialNumber_", d.serialNumber, -1)
	*pc = strings.Replace(*pc, "_Weight_", d.weight, -1)
	*pc = strings.Replace(*pc, "_Month_", month, -1)
	*pc = strings.Replace(*pc, "_Year_", year, -1)
	*pc = strings.Replace(*pc, "_RevLevel_", d.revLevel, -1)
}

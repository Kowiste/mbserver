package main

import (
	"log"
	"net"
	"time"

	md "github.com/Kowiste/modserver"
)

func main() {

	serv := md.NewServer()
	serv.RegisterFunctionHandler(3, CustomHandler)
	serv.OnConnectionHandler(ConnectionHandler)

	err := serv.ListenTCP("0.0.0.0:40502")
	if err != nil {
		log.Printf("%v\n", err)
	}
	defer serv.Close()
	log.Println("Server Active")

	// Wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}

//CustomHandler fucntion to customize the server response
func CustomHandler(s *md.Server, frame md.Framer) ([]byte, *md.Exception) {
	_, numRegs, endRegister := frame.RegisterAddressAndNumber(frame)
	// Check the request is within the allocated memory
	if endRegister > 65535 {
		return []byte{}, &md.IllegalDataAddress
	}
	dataSize := numRegs / 8
	data := make([]byte, 1+dataSize)
	data[0] = byte(dataSize)
	return data, &md.Success
}

//ConnectionHandler On connection
func ConnectionHandler(IP net.Addr) {
	log.Println("Connection Establish from: ", IP.String())
}

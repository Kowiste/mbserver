package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	md "github.com/Kowiste/modserver"
)

var memory []uint16

func main() {
	port := flag.String("p", "40502", "Port to deploy Modbus")
	mem := flag.String("mem", "", "Path to the configuration memory json")
	mode := flag.Int("m", 3, "Mode of the server:	1 = ReadCoils, 2 = ReadDiscreteInputs, 3 = ReadHoldingRegisters, 4 = ReadInputRegisters, 5 = WriteSingleCoil, 6 = WriteHoldingRegister,	15 = WriteMultipleCoils, 16 = WriteHoldingRegisters ")

	flag.Parse()
	if *mem != "" {
		memory = loadMemory(*mem)
	}
	serv := md.NewServer()
	serv.HoldingRegisters = memory
	if *mode !=0{
		serv.RegisterFunctionHandler(uint8(*mode), CustomHandler)
	}
	serv.OnConnectionHandler(ConnectionHandler)
	serv.OnTimerHandler(TimerHandler, 1*time.Second)

	err := serv.ListenTCP("0.0.0.0:" + *port)
	if err != nil {
		log.Printf("%v\n", err)
	}
	defer serv.Close()
	log.Println("[Author: kowiste] Modbus Server Active on port", *port)

	// Wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}

//CustomHandler fucntion to customize the server response
func CustomHandler(s *md.Server, frame md.Framer) ([]byte, *md.Exception) {
	reg, numRegs, _ := frame.RegisterAddressAndNumber(frame)
	data := make([]byte, numRegs+1)
	data[0] = byte(numRegs) //the number of byte to send
	dataPointer := 1        //Pointer of the first valid elemet in the array
	if len(memory) >= reg+(numRegs/2) {
		for n := 0; n < numRegs/2; n++ {
			num := int16ToByte(memory[reg+n])
			data[dataPointer] = num[0]
			data[dataPointer+1] = num[1]
			dataPointer += 2
		}
		log.Println("Reading Address: " + strconv.Itoa(reg) + " reading " + strconv.Itoa(numRegs) + " bytes")
		return data, &md.Success
	}
	log.Println("Illegal Address")
	return data, &md.IllegalDataAddress //return illegal addresss
}

//ConnectionHandler On connection
func ConnectionHandler(IP net.Addr) {
	log.Println("Connection Establish from: ", IP.String())
}

//TimerHandler on timer handler pout the code you want to execute every time given
func TimerHandler(s *md.Server) {
	log.Println("Timer happen")
	s.HoldingRegisters[2]++
}

func loadMemory(path string) []uint16 {
	mem := make([]uint16, 0)
	b, err := ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &mem)
		if err != nil {
			println(err.Error())
			return nil
		}
		return mem
	}
	return nil
}

//ReadFile Read a File
func ReadFile(FilePath string) ([]byte, error) {
	file, err := os.Open(FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func int16ToByte(input uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, input)
	return b
}

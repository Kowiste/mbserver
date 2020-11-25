package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"strconv"
	"time"

	md "github.com/Kowiste/modserver"
	"github.com/kowiste/utils/conversion/array"
	"github.com/kowiste/utils/conversion/number"
	"github.com/kowiste/utils/file"
	"github.com/kowiste/utils/generator/location"
	plc "github.com/kowiste/utils/plc/generate/location"
	"github.com/kowiste/utils/plc/generate/other"
)

var memory []uint16
var sec int
var msgCount uint16
var text = []string{"OK", "NG", "Lazada Ad.", "Check In"}
var geo *location.GeoLocationHelper

func main() {
	port := flag.String("p", "40102", "Port to deploy Modbus")
	mem := flag.String("mem", "", "Path to the configuration memory json")
	mode := flag.Int("m", 3, "Mode of the server:	1 = ReadCoils, 2 = ReadDiscreteInputs, 3 = ReadHoldingRegisters, 4 = ReadInputRegisters, 5 = WriteSingleCoil, 6 = WriteHoldingRegister,	15 = WriteMultipleCoils, 16 = WriteHoldingRegisters ")
	tick := flag.Int("t", 0, "Millisecond to trigger ontimer")

	flag.Parse()
	geo = location.NewGeoLocnRnd(0.01)
	b, _ := file.Read("device.json")
	geo.LoadnoZ(b)

	serv := md.NewServer()
	if *mem != "" {
		memory = loadMemory(*mem)
		serv.HoldingRegisters = memory
	} else {
		serv.HoldingRegisters = make([]uint16, ^uint16(0))
	}

	if *mode != 0 {
		//serv.RegisterFunctionHandler(uint8(*mode), CustomHandler)
	}
	serv.OnConnectionHandler(ConnectionHandler)
	if *tick != 0 {
		serv.OnTimerHandler(TimerHandler, time.Duration(*tick)*time.Millisecond)
	}

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
			num := number.Uint16ToByteArr(memory[reg+n])
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
	data := array.ByteToUint16Arr(loadDevice(), true)
	for index := range data {
		s.HoldingRegisters[index] = data[index]
	}
}

func loadMemory(path string) []uint16 {
	mem := make([]uint16, 0)
	b, err := file.Read(path)
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

func loadDevice() []byte {
	index := 0
	out := make([]byte, 38)
	///////////////////////
	// Loading data Arduino
	///////////////////////
	out[index] = 0
	out[index+1] = 1
	status := other.RandomBool()
	if sec%17 == 0 { //Every 17 second
		if !status { //Connection Status [0]
			out[index+1] = 0
		}
	}
	index += 2
	//message count[1]
	msgCount += uint16(other.RandomInt(23445, 1))
	numMsg := number.Uint16ToByteArr(msgCount)
	for element := range numMsg {
		out[index+element] = numMsg[element]
	}
	index += len(numMsg)

	//link quality[2]
	lq := other.RandomFloat(70, 0)
	linkQ := number.Float64ToByteArr(lq)
	for element := range linkQ {
		out[index+element] = linkQ[element]
	}
	index += len(linkQ)

	//geo position[6]
	b := plc.ConvLocToByteArr(geo.Actual, false)
	for element := range b {
		out[element+index] = b[element]
	}
	index += len(b)

	//string [14] len(10) bytes
	byteText := []byte(text[other.RandomInt(4, 0)])
	for element := range byteText {
		out[element+index] = byteText[element]
	}
	index += len(byteText)
	for i := index; i < len(out); i++ {
		out[i] = 32
	}
	println("status: ", status, " Cnt: ", msgCount, " Link Quality: ", lq, " geo: ", geo.Actual.Latitude, " ,", geo.Actual.Longitude, " text: ", string(byteText))
	sec++
	geo.Next() //updating position
	return out
}

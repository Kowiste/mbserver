package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"time"

	md "github.com/Kowiste/modserver"
)

var memory []uint16
var geoMap map[string][]float64
var GTruck GroupTrucks

type GroupTrucks struct {
	Trucks map[string]*Truck
}
type Truck struct {
	Max  int
	Next int
	GPS  []float64
}

//var
func main() {
	port := flag.String("p", "40502", "Port to deploy Modbus")
	mem := flag.String("mem", "", "Path to the configuration memory json")
	//mode := flag.Int("m", 3, "Mode of the server:	1 = ReadCoils, 2 = ReadDiscreteInputs, 3 = ReadHoldingRegisters, 4 = ReadInputRegisters, 5 = WriteSingleCoil, 6 = WriteHoldingRegister,	15 = WriteMultipleCoils, 16 = WriteHoldingRegisters ")
	tick := flag.Int("t", 0, "Millisecond to trigger ontimer")
	GTruck.Trucks = make(map[string]*Truck)

	flag.Parse()
	if *mem != "" {
		//memory =  //loadMemory(*mem)
		memory = make([]uint16, ^uint16(0))
		geoMap = loadGeo(*mem)
		for element := range geoMap {
			GTruck.Trucks[element] = &Truck{Max: len(geoMap[element]), GPS: geoMap[element]}
		}
	}
	GTruck.updateGeo()
	serv := md.NewServer()
	serv.HoldingRegisters = memory
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

//ConnectionHandler On connection
func ConnectionHandler(IP net.Addr) {
	log.Println("Connection Establish from: ", IP.String())
}

//TimerHandler on timer handler pout the code you want to execute every time given
func TimerHandler(s *md.Server) {
	//log.Println("Updating values")
	GTruck.updateGeo()
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
func loadGeo(path string) map[string][]float64 {
	mem := make(map[string][]float64)
	b, err := ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &mem)
		if err != nil {
			println(err.Error())
			return nil
		}
	}

	return mem
}

func doubleToByte(input []float64, pointer int) []uint16 {
	buf := make([]uint16, 0)
	n := math.Float64bits(input[pointer])
	buf = append(buf, uint16(n>>48))
	buf = append(buf, uint16(n>>32))
	buf = append(buf, uint16(n>>16))
	buf = append(buf, uint16(n))
	n = math.Float64bits(input[pointer+1])
	buf = append(buf, uint16(n>>48))
	buf = append(buf, uint16(n>>32))
	buf = append(buf, uint16(n>>16))
	buf = append(buf, uint16(n))
	return buf
}

func (t *GroupTrucks) updateGeo() {
	for truck := range t.Trucks {
		retMem := doubleToByte(geoMap[truck], t.Trucks[truck].Next) //Get the next 2 float for every truck
		if t.Trucks[truck].Next < t.Trucks[truck].Max-2 {           //Check for the limit
			t.Trucks[truck].Next += 2 //Pointer to the next element 2 position
		} else {
			t.Trucks[truck].Next = 0
		}

		num, _ := strconv.Atoi(truck)
		for element := range retMem {
			memory[num+element] = retMem[element]
		}
	}
}

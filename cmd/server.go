package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net"
	"time"

	md "github.com/Kowiste/modserver"
)

//constant
const (
	NTimesActive int = 20
)

var rdSource rand.Source
var rd *rand.Rand
var memory []uint16
var times byte

func main() {
	port := flag.String("p", "40502", "Port to deploy Modbus")
	mem := flag.String("mem", "", "Path to the configuration memory json")
	tick := flag.Int("t", 0, "Millisecond to trigger ontimer")
	GTruck.Trucks = make(map[string]*Truck)
	rdSource = rand.NewSource(time.Now().UnixNano())
	rd = rand.New(rdSource)
	flag.Parse()
	if *mem != "" {
		//memory =  //loadMemory(*mem)
		memory = make([]uint16, ^uint16(0))
		geoMap = loadGeo(*mem)
		element := "1000"
		//for element := range geoMap {
		GTruck.Trucks[element] = &Truck{Max: len(geoMap[element]), GPS: geoMap[element]}
		GTruck.Trucks[element].pumps = loadPump()
		GTruck.Trucks[element].TimesAct = NTimesActive
		//}
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
	if times == 2 {
		GTruck.updateGeo()
		times = 0
	}
	GTruck.updatePump()
	times++
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
func loadPump() map[byte]*Pump {
	t := make(map[byte]*Pump)
	for i := 0; i < 2; i++ {
		t[byte(i)] = NewPump(380, 380, 30, 5, 2.5)
	}
	return t
}
 
// Package mbserver implments a Modbus server (slave).
package mbserver

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/goburrow/serial"
	log "github.com/sirupsen/logrus"
)

// Server is a Modbus slave with allocated memory for discrete inputs, coils, etc.
type Server struct {
	// Debug enables more verbose messaging.
	Debug            bool
	listeners        []net.Listener
	ports            []serial.Port
	requestChan      chan *Request
	function         [256](func(*Server, Framer) ([]byte, *Exception))
	onConnection     (func(net.Addr))
	onTimer          (func(*Server))
	tick             time.Duration
	DiscreteInputs   []byte
	Coils            []byte
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

// Request contains the connection and Modbus frame.
type Request struct {
	conn  io.ReadWriteCloser
	frame Framer
}

// NewServer creates a new Modbus server (slave).
func NewServer() *Server {
	s := &Server{}

	// Allocate Modbus memory maps.
	s.DiscreteInputs = make([]byte, 65536)
	s.Coils = make([]byte, 65536)
	s.HoldingRegisters = make([]uint16, 65536)
	s.InputRegisters = make([]uint16, 65536)

	// Add default functions.
	s.function[1] = ReadCoils
	s.function[2] = ReadDiscreteInputs
	s.function[3] = ReadHoldingRegisters
	s.function[4] = ReadInputRegisters
	s.function[5] = WriteSingleCoil
	s.function[6] = WriteHoldingRegister
	s.function[15] = WriteMultipleCoils
	s.function[16] = WriteHoldingRegisters

	s.requestChan = make(chan *Request)
	go s.handler()

	return s
}

// RegisterFunctionHandler override the default behavior for a given Modbus function.
func (s *Server) RegisterFunctionHandler(funcCode uint8, function func(*Server, Framer) ([]byte, *Exception)) {
	s.function[funcCode] = function
}

//OnTimerHandler Function that happend when there is a new conection
func (s *Server) OnTimerHandler(function func(*Server), tick time.Duration) {
	s.onTimer = function
	s.tick = tick
	go s.timing(s.tick)
}

//OnConnectionHandler Function that happend when there is a new conection
func (s *Server) OnConnectionHandler(function func(net.Addr)) {
	s.onConnection = function
}

func (s *Server) handle(request *Request) Framer {
	var exception *Exception
	var data []byte

	response := request.frame.Copy()

	function := request.frame.GetFunction()
	if s.function[function] != nil {
		data, exception = s.function[function](s, request.frame) //Call the fucntion when there is request
		response.SetData(data)
	} else {
		exception = &IllegalFunction
		log.Error("Exception handler Illegal function")
	}

	if exception != &Success {
		response.SetException(exception)
	}

	return response
}

// All requests are handled synchronously to prevent modbus memory corruption.
func (s *Server) handler() {
	for {
		request := <-s.requestChan
		response := s.handle(request)
		fmt.Println("%h", response.Bytes())
		request.conn.Write(response.Bytes())
	}
}

func (s *Server) timing(tick time.Duration) {
	t := time.NewTicker(tick)
	for {
		<-t.C //Wait for the time
		s.onTimer(s)
	}
}

// Close stops listening to TCP/IP ports and closes serial ports.
func (s *Server) Close() {
	for _, listen := range s.listeners {
		listen.Close()
	}
	for _, port := range s.ports {
		port.Close()
	}
}

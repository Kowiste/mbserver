package mbserver

import (
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Error("Unable to accept connections: %#v\n", err)
			return err
		}
		if s.onConnection != nil {
			go s.onConnection(conn.RemoteAddr())
		}
		go func(conn net.Conn) {
			defer conn.Close()

			for {
				packet := make([]byte, 512)
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						log.Error("read error %v\n", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
					log.Error("bad packet error %v\n", err)
					return
				}

				request := &Request{conn, frame}

				s.requestChan <- request
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Error("Failed to Listen: %v\n", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}

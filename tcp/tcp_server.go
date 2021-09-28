package tcp

import (
	"fmt"
	"net"
	"user-management/log"
	message "user-management/message"

	"google.golang.org/protobuf/proto"
)

type ListenCallback func(net.Conn) error

type TCPServer struct {
	Address        string
	Listener       net.Listener
	HandleCallback ListenCallback
}

func NewTCPServer(address string, callback ListenCallback) (*TCPServer, error) {
	TCPsv := &TCPServer{
		Address:        address,
		HandleCallback: callback,
	}
	return TCPsv, nil
}

func (tcp *TCPServer) SendTCPData(connection net.Conn, mess []byte) {
	err := SendTCPData(connection, mess)
	if err != nil {
		connection.Close()
	}
}

func (tcp *TCPServer) ReadTCPData(c net.Conn) (*message.MessageRequest, error) {
	messBytes, err := ReadTCPData(c)
	if err != nil {
		c.Close()
		return nil, err
	}
	messReq := message.MessageRequest{}
	err = proto.Unmarshal(messBytes, &messReq)
	if err != nil {
		c.Close()
		return nil, err
	}
	return &messReq, nil
}

func (tcp *TCPServer) Start() error {
	log.Log.InfoLogger.Println("TCP server start listener: " + tcp.Address)
	l, err := net.Listen("tcp", tcp.Address)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tcp.Listener = l
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go tcp.HandleCallback(c)
	}
}

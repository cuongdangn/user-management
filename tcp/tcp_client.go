package tcp

import (
	"errors"
	"net"
	"sync"
	"time"
	"user-management/log"
)

var (
	ErrConnections = errors.New("Number of connection reached limit")
	ErrDial        = errors.New("Dial fail")
	ErrTimeOut     = errors.New("Get Connection timeout")
)

type TCPClient struct {
	Address       string
	Connections   chan net.Conn
	NumConnection int
	MaxConnection int
	Lock          sync.Mutex
}

func NewTCPClient(address string, maxConection int) (*TCPClient, error) {
	TCPPool := &TCPClient{
		Address:       address,
		MaxConnection: maxConection,
		NumConnection: 0,
		Connections:   make(chan net.Conn, maxConection),
	}
	//TCPPool.initConnection(maxConection / 5)
	return TCPPool, nil
}

func (tcp *TCPClient) initConnection(numConnection int) {
	log.Log.InfoLogger.Println("Init the connections")
	for i := 0; i < numConnection; i++ {
		conn, err := net.Dial("tcp", tcp.Address)
		if err == nil {
			tcp.PutConnection(conn)
		} else {
			log.Log.ErrorLogger.Println(err)
		}
	}
}

func (tcp *TCPClient) SendTCPData(mess []byte) (net.Conn, error) {
	connection, err := tcp.getConnection()

	if err != nil {
		tcp.DeleteConnection(connection)
		return nil, err
	}

	err = SendTCPData(connection, mess)
	if err != nil {
		tcp.DeleteConnection(connection)
		return nil, err
	}
	return connection, nil
}

func (tcp *TCPClient) ReadTCPData(c net.Conn) ([]byte, error) {
	mess, err := ReadTCPData(c)
	if err != nil {
		tcp.DeleteConnection(c)
		return nil, err
	}
	return mess, err
}

func (tcp *TCPClient) getConnection() (net.Conn, error) {
	timeOut := time.After(time.Duration(1) * time.Second)
	for {
		select {
		case conn := <-tcp.Connections:
			return conn, nil
		case <-timeOut:
			return nil, ErrTimeOut
		default:
			conn, err := tcp.createConn()
			if err == nil {
				return conn, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (tcp *TCPClient) PutConnection(c net.Conn) {
	tcp.Connections <- c
}

func (tcp *TCPClient) DeleteConnection(c net.Conn) {
	if c == nil {
		return
	}
	tcp.Lock.Lock()
	defer tcp.Lock.Unlock()
	tcp.NumConnection -= 1
	c.Close()
}

func (tcp *TCPClient) createConn() (net.Conn, error) {
	tcp.Lock.Lock()
	defer tcp.Lock.Unlock()
	if tcp.NumConnection >= tcp.MaxConnection {
		return nil, ErrConnections
	}
	conn, err := net.Dial("tcp", tcp.Address)
	if err != nil {
		log.Log.ErrorLogger.Println(ErrDial.Error())
		return nil, ErrDial
	}
	tcp.NumConnection = tcp.NumConnection + 1
	return conn, nil
}

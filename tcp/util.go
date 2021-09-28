package tcp

import (
	"encoding/binary"
	"io"
	"net"
)

func FillString(crStr string, targetLength int) string {
	crlen := len(crStr)
	for i := crlen; i < targetLength; i++ {
		crStr += ":"
	}
	return crStr
}

func SendTCPData(connection net.Conn, mess []byte) error {
	var sizeMessage = len(mess)
	buff := make([]byte, 8)
	binary.PutVarint(buff, int64(sizeMessage))
	_, err := connection.Write(buff)
	if err != nil {
		return err
	}
	_, err = connection.Write(mess)
	if err != nil {
		return err
	}

	return nil
}

func ReadTCPData(c net.Conn) ([]byte, error) {
	bufferMessSize := make([]byte, 8)
	_, err := c.Read(bufferMessSize)

	if err != nil {
		return nil, err
	}

	messSize, _ := binary.Varint(bufferMessSize)
	bufferMess := make([]byte, messSize)
	_, err = io.ReadFull(c, bufferMess)

	if err != nil {
		return nil, err
	}

	return bufferMess, nil
}

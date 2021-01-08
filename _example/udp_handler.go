package _example

import (
	"fmt"
	"github.com/ilooky/mixer"
	"net"
)

type UdpHandler struct {
	udpConn *net.UDPConn
}

func (u UdpHandler) Send(data []byte) error {
	size, err := u.udpConn.Write(data)
	if err != nil {
		return err
	}
	fmt.Println("send size=", size)
	return nil
}

func (u UdpHandler) Read() error {
	data := make([]byte, mixer.PckSize)
	n, remoteAddr, err := u.udpConn.ReadFromUDP(data) // 接收数据
	if err != nil {
		fmt.Println("接收数据失败，err:", err)
		return err
	}

	fmt.Printf("recv:%v addr:%v count:%v\n", string(data[:n]), remoteAddr, n)
	return nil
}

func (u UdpHandler) Close() {
	u.udpConn.Close()
}

func NewUdpHandler() mixer.IHandler {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 30000,
	})
	if err != nil {
		panic(err)
	}
	return &UdpHandler{
		udpConn: socket,
	}
}

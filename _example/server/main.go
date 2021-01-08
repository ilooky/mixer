package main

import (
	"github.com/ilooky/logger"
	"github.com/ilooky/mixer"
	"github.com/ilooky/mixer/_example"
	"net"
)

func main() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 30000,
	})
	if err != nil {
		logger.Error("listen failed, err:", err)
		return
	}
	defer listen.Close()
	udpHandler := _example.NewUdpHandler()
	mix := mixer.NewMixer(mixer.Config{
		Handler: udpHandler,
	})
	for {
		var data [mixer.PckSize]byte
		n, addr, err := listen.ReadFromUDP(data[:]) // 接收数据
		if err != nil {
			logger.Error("read udp failed, err:", err)
			continue
		}
		id := string(data[0:1])
		logger.InfoKV("pack", "id", id)
		index := string(data[1:3])
		logger.InfoKV("pack", "index", index)
		flag := string(data[n-1 : n])
		logger.InfoKV("pack", "flag", flag)
		content := string(data[3 : n-1])
		logger.InfoKV("pack", "content", content)
		err = mix.SaveRecv(id, index, flag, data[3:n-1])
		if err != nil {
			logger.Error(err)
		}
		if flag == "1" {
			message, err := mix.GetRecvById(id, index)
			if err != nil {
				logger.Error(err)
			} else {
				logger.Info(message)
			}
		}
		_, err = listen.WriteToUDP(data[:n], addr) // 发送数据
		if err != nil {
			logger.Error("write to udp failed, err:", err)
			continue
		}
	}
}

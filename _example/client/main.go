package main

import (
	"bufio"
	"github.com/ilooky/logger"
	"github.com/ilooky/mixer"
	"github.com/ilooky/mixer/_example"
	"os"
)

// UDP 客户端
func main() {
	udpHandler := _example.NewUdpHandler()
	mix := mixer.NewMixer(mixer.Config{
		Handler: udpHandler,
	})
	defer mix.Close()
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				logger.Error(err)
			}
			err = mix.Submit(line)
			if err != nil {
				logger.Error(err)
			}
		}
	}()
	mix.Run()

}

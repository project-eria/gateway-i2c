package main

import (
	eria "github.com/project-eria/eria-core"
	"gobot.io/x/gobot/platforms/raspi"
)

type i2cDevice interface {
	config(int, *eria.EriaServer, *raspi.Adaptor, string) *eria.EriaThing
	runLoop()
}

func runLoops() {
	for _, oDevice := range _devices {
		go oDevice.runLoop()
	}
}

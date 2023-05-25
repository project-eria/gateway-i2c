package main

import (
	"fmt"

	eria "github.com/project-eria/eria-core"
	"github.com/rs/zerolog/log"
	"gobot.io/x/gobot/platforms/raspi"
)

var config = struct {
	Host        string         `yaml:"host"`
	Port        uint           `yaml:"port" default:"80"`
	ExposedAddr string         `yaml:"exposedAddr"`
	ID          string         `yaml:"id" required:"true"`
	Devices     []configDevice `yaml:"devices"`
}{}

type configDevice struct {
	Type   string                 `yaml:"type" required:"true"`
	Rate   int                    `yaml:"rate"` // seconds
	Params map[string]interface{} `yaml:"params"`
}

var (
	_devices = []i2cDevice{}
)

func main() {
	defer func() {
		log.Info().Msg("[main] Stopped")
	}()

	eria.Init("ERIA i2c Gateway")

	// Loading config
	eria.LoadConfig(&config)

	eriaServer := eria.NewServer(config.Host, config.Port, config.ExposedAddr, "")

	board := raspi.NewAdaptor()

	for i := range config.Devices {
		confDev := &config.Devices[i]
		var iDevice i2cDevice
		switch confDev.Type {
		case "ads1115":
			iDevice = newADS1115(confDev)
		case "sht31d":
			iDevice = newSHT31D(confDev)
		case "bmp388":
			iDevice = newBMP388(confDev)
		default:
			log.Warn().Str("type", confDev.Type).Msg("[main] Not supported i2c type")
		}
		if iDevice != nil {
			urn := fmt.Sprintf("eria:gateway:i2c:%s:%d", config.ID, i)
			iDevice.config(i, eriaServer, board, urn)
			_devices = append(_devices, iDevice)
		}
	}

	runLoops()

	eriaServer.StartServer()
}

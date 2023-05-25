package main

import (
	"fmt"
	"time"

	eria "github.com/project-eria/eria-core"
	"github.com/project-eria/go-wot/producer"
	"github.com/rs/zerolog/log"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

type bmp388 struct {
	driver *i2c.BMP388Driver
	iThing *eria.EriaThing
	*configDevice
}

func newBMP388(confDev *configDevice) i2cDevice {
	return &bmp388{
		configDevice: confDev,
	}
}

func (d *bmp388) config(i int, eriaServer *eria.EriaServer, board *raspi.Adaptor, urn string) *eria.EriaThing {
	td, _ := eria.NewThingDescription(
		urn,
		eria.AppVersion,
		"bmp388",
		"i2c bmp388",
		[]string{"TemperatureSensor", "BarometerSensor"},
	)
	td.Properties["temperature"].ReadOnly = true
	td.Properties["pressure"].ReadOnly = true
	if d.Rate == 0 {
		td.Properties["temperature"].Observable = false
		td.Properties["pressure"].Observable = false
	}
	d.driver = i2c.NewBMP388Driver(board,
		i2c.WithBus(1),
		i2c.WithAddress(0x77))

	if err := d.driver.Start(); err != nil {
		log.Fatal().Err(err).Msg("[main:bmp388] Start")
	}
	d.iThing, _ = eriaServer.AddThing(fmt.Sprintf("device.%d", i), td)

	d.iThing.SetPropertyReadHandler("temperature", func(t *producer.ExposedThing, name string) (interface{}, error) {
		return d.getTemperatureValue()
	})
	d.iThing.SetPropertyReadHandler("pressure", func(t *producer.ExposedThing, name string) (interface{}, error) {
		return d.getPressureValue()
	})
	return d.iThing
}

func (d *bmp388) getTemperatureValue() (interface{}, error) {
	temperature, err := d.driver.Temperature(i2c.BMP388AccuracyLow)
	if err != nil {
		log.Error().Err(err).Msg("[main:bmp388] Temperature Sample")
	} else {
		log.Info().
			Float32("temperature", temperature).
			Msg("[main:bmp388] Data requested")
	}
	return temperature, err
}

func (d *bmp388) getPressureValue() (interface{}, error) {
	pressure, err := d.driver.Pressure(i2c.BMP388AccuracyHigh)
	if err != nil {
		log.Error().Err(err).Msg("[main:bmp388] Pressure Sample")
	} else {

		// Convert to hPa
		pressure = pressure / 100

		log.Info().
			Float32("pressure", pressure).
			Msg("[main:bmp388] Sending i2c update")

	}
	return pressure, err
}

func (d *bmp388) update() {
	log.Trace().Msg("[main:bmp388] Regular value update")
	temperature, _ := d.getTemperatureValue()
	d.iThing.SetPropertyValue("temperature", temperature)
	pressure, _ := d.getPressureValue()
	d.iThing.SetPropertyValue("pressure", pressure)

}

func (d *bmp388) runLoop() {
	if d.Rate > 0 {
		ticker := time.NewTicker(time.Duration(d.Rate) * time.Second)
		d.update()
		for {
			<-ticker.C
			d.update()
		}
	}
}

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

type sht31d struct {
	driver *i2c.SHT3xDriver
	iThing *eria.EriaThing
	*configDevice
}

func newSHT31D(confDev *configDevice) i2cDevice {
	return &sht31d{
		configDevice: confDev,
	}
}

func (d *sht31d) config(i int, eriaServer *eria.EriaServer, board *raspi.Adaptor, urn string) *eria.EriaThing {
	td, _ := eria.NewThingDescription(
		urn,
		eria.AppVersion,
		"sht31d",
		"i2c sht31d",
		[]string{"TemperatureSensor", "HygrometerSensor"},
	)
	td.Properties["temperature"].ReadOnly = true
	td.Properties["humidity"].ReadOnly = true
	if d.Rate == 0 {
		td.Properties["temperature"].Observable = false
		td.Properties["humidity"].Observable = false
	}

	d.driver = i2c.NewSHT3xDriver(board)

	if err := d.driver.Start(); err != nil {
		log.Fatal().Err(err).Msg("[main:sht31d] Start")
	}
	d.iThing, _ = eriaServer.AddThing(fmt.Sprintf("device.%d", i), td)

	d.iThing.SetPropertyReadHandler("temperature", func(t *producer.ExposedThing, name string) (interface{}, error) {
		return d.getTemperatureValue()
	})
	d.iThing.SetPropertyReadHandler("humidity", func(t *producer.ExposedThing, name string) (interface{}, error) {
		return d.getHumidityValue()
	})
	return d.iThing
}

func (d *sht31d) getTemperatureValue() (interface{}, error) {
	temperature, _, err := d.driver.Sample()
	if err != nil {
		log.Fatal().Err(err).Msg("[main:sht31d] Temperature Samples")
	} else {
		log.Info().
			Float32("temperature", temperature).
			Msg("[main:sht31d] data requested")
	}

	return temperature, err
}

func (d *sht31d) getHumidityValue() (interface{}, error) {
	_, rh, err := d.driver.Sample()
	if err != nil {
		log.Fatal().Err(err).Msg("[main:sht31d] Humidity Samples")
	} else {
		log.Info().
			Float32("rh", rh).
			Msg("[main:sht31d] data requested")
	}

	return rh, err
}

func (d *sht31d) update() {
	log.Trace().Msg("[main:sht31d] Regular value update")
	temperature, _ := d.getTemperatureValue()
	d.iThing.SetPropertyValue("temperature", temperature)
	humidity, _ := d.getHumidityValue()
	d.iThing.SetPropertyValue("humidity", humidity)
}

func (d *sht31d) runLoop() {
	if d.Rate > 0 {
		ticker := time.NewTicker(time.Duration(d.Rate) * time.Second)
		d.update()
		for {
			<-ticker.C
			d.update()
		}
	}
}

package main

import (
	"fmt"
	"strconv"
	"time"

	eria "github.com/project-eria/eria-core"
	"github.com/project-eria/go-wot/producer"
	"github.com/rs/zerolog/log"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

/* https://gobot.io/documentation/drivers/ads1115/ */

type ads1115 struct {
	driver *i2c.ADS1x15Driver
	iThing *eria.EriaThing
	gains  map[string]int
	*configDevice
}

func newADS1115(confDev *configDevice) i2cDevice {
	return &ads1115{
		gains:        map[string]int{},
		configDevice: confDev,
	}
}

func (d *ads1115) config(i int, eriaServer *eria.EriaServer, board *raspi.Adaptor, urn string) *eria.EriaThing {
	td, _ := eria.NewThingDescription(
		urn,
		eria.AppVersion,
		"ads1115",
		"i2c ads1115",
		[]string{},
	)

	d.driver = i2c.NewADS1115Driver(board)

	for channel, inputType := range d.Params {
		switch inputType {
		case "volts":
			// Adjust the gain to be able to read values of at least 5V
			d.gains[channel], _ = d.driver.BestGainForVoltage(5.0)
			gainVoltage := map[int]float64{
				2 / 3: 6.144,
				1:     4.096,
				2:     2.048,
				4:     1.024,
				8:     0.512,
				16:    0.256,
			}
			log.Info().Float64("gain", gainVoltage[d.gains[channel]]).Msg("[main:ads1115] Gain")
			postfix := fmt.Sprintf(".%s", channel)
			if err := eria.AddModel(td, "VoltageSensor", postfix); err != nil {
				log.Fatal().Err(err).Msg("[main:ads1115] AddSchema")
			}
			td.Properties["volts"+postfix].ReadOnly = true
			if d.Rate == 0 {
				td.Properties["volts"+postfix].Observable = false
			}
		default:
			log.Warn().Str("type", d.Type).Msg("[main:ads1115] Unknown input type")
		}
	}

	if err := d.driver.Start(); err != nil {
		log.Fatal().Err(err).Msg("[main:ads1115] Start")
	}

	d.iThing, _ = eriaServer.AddThing(fmt.Sprintf("device.%d", i), td)
	for channel, inputType := range d.Params {
		switch inputType {
		case "volts":
			postfix := fmt.Sprintf(".%s", channel)
			d.iThing.SetPropertyReadHandler("volts"+postfix, func(t *producer.ExposedThing, name string) (interface{}, error) {
				return d.getValue(channel)
			})
		}
	}
	return d.iThing
}

func (d *ads1115) getValue(channel string) (interface{}, error) {
	addr, _ := strconv.Atoi(channel)
	value, err := d.driver.Read(addr, d.gains[channel], 128)
	if err != nil {
		log.Fatal().Err(err).Msg("[main] ReadWithDefaults")
	}

	log.Trace().
		Float64(d.Type, value).
		Str("i2c Channel", channel).
		Msg("[main] Get sensor data")
	return value, err
}

func (d *ads1115) update() {
	log.Trace().Msg("[main:ads1115] Regular value update")

	for channel, inputType := range d.Params {
		switch inputType {
		case "volts":
			property := fmt.Sprintf("volts.%s", channel)
			value, _ := d.getValue(channel)
			d.iThing.SetPropertyValue(property, value)
		default:
			log.Warn().Str("type", d.Type).Msg("[main:ads1115] Unknown input type")
		}
	}
}

func (d *ads1115) runLoop() {
	if d.Rate > 0 {
		ticker := time.NewTicker(time.Duration(d.Rate) * time.Second)
		d.update()
		for {
			<-ticker.C
			d.update()
		}
	}
}

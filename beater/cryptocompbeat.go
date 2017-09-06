package beater

import (
	"fmt"
	"strings"
	//	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/gravewrought/cryptocompbeat/config"
)

type Cryptocompbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Cryptocompbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Cryptocompbeat) Run(b *beat.Beat) error {
	logp.Info("cryptocompbeat is running! Hit CTRL-C to stop it.")

	var err error

	ccs := Cryptocompsub{
		Type:       "0",
		Exchange:   "BitTrex",
		SymbolFrom: "BTC",
		SymbolTo:   "USD",
		Data:       make(chan Cryptocompdata),
	}

	logp.Info("Sub Open")
	ccs.Open()

	//bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	logp.Info("Run Start")

	for {
		select {
		case <-bt.done:
			return nil

		case data := <-ccs.Data:
			logp.Info(data.Msg)

			switch data.Type {
			case 0:
				break

			case 1:
				parsed := strings.Split(data.Msg, "~")

				if parsed[0] == "0" {
					//			logp.Info("timestamp " + time.Now())
					logp.Info("exchange " + parsed[1])
					logp.Info("from_symbol " + parsed[2])
					logp.Info("to_symbol " + parsed[3])
					logp.Info("rate " + parsed[8])
				}

				/*
					parsed := common.MapStr{
						"exchange":    parsed[1],
						"from_symbol": parsed[7],
						"to_symbol":   parsed[8],
						"rate":        parsed[3],
					}
								event := beat.Event{
									Timestamp: time.Now(),
									Fields: parsed,
								}
					logp.Info(parsed)
					/**/
				break
			}
		}
		/*
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":    b.Info.Name,
					"counter": counter,
				},
			}
			bt.client.Publish(event)
			/**/
	}

	logp.Info("Run End")
	ccs.Close()

	return nil
}

func (bt *Cryptocompbeat) Stop() {
	//bt.client.Close()
	close(bt.done)
}

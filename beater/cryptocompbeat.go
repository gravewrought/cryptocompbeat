package beater

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	//
	// Create the Cryptocompsub object -- this object manages the subscription
	// with CryptoCurrency.
	//
	ccs := Cryptocompsub{
		Type:       "0",
		Exchange:   bt.config.Exchange,
		SymbolFrom: bt.config.SymbolFrom,
		SymbolTo:   bt.config.SymbolTo,
		Data:       make(chan Cryptocompdata),
	}

	//
	// Open the connection to CryptoCurrency
	//
	ccs.Open()

	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	//
	// Running Loop
	//
	for {
		//
		// Wait for channels
		//
		select {
		// Check if the Cryptocompbeat has been closed
		case <-bt.done:
			return nil

		// Check if data has been received by the Cryptocombsub
		case data := <-ccs.Data:
			switch data.Type {
			case 0:
				break

			case 1:
				//
				// Split the message on a tilda delimiter
				//
				parsed := strings.Split(data.Msg, "~")

				//
				// Parse the code in the message, which affects the
				// size of the parsed message that was received.
				//
				code, _ := strconv.Atoi(parsed[0])

				switch code {
				//
				// Zero is a message with transaction data
				//
				case 0:
					rate, _ := strconv.ParseFloat(parsed[8], 32)

					logp.Info(fmt.Sprintf(
						"%s -> %s %f (%s)",
						parsed[2],
						parsed[3],
						rate,
						parsed[1],
					))

					parsed := common.MapStr{
						"exchange":    parsed[1],
						"symbol_from": parsed[2],
						"symbol_to":   parsed[3],
						"rate":        rate,
					}

					event := beat.Event{
						Timestamp: time.Now(),
						Fields:    parsed,
					}

					bt.client.Publish(event)

					break

				//
				// By default assume the message is informative, print
				// it to the log.
				//
				default:
					logp.Info("Message: " + data.Msg)
					break
				}

				break
			}
		}
	}

	//
	// Close the connection to CryptoCurrency
	//
	ccs.Close()

	return nil
}

func (bt *Cryptocompbeat) Stop() {
	//bt.client.Close()
	close(bt.done)
}

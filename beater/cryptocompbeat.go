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
		Type:       bt.config.Type,
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
				// Print the message!
				//
				logp.Info("Message: " + data.Msg)

				//
				// Split the message on a tilda delimiter
				//
				split := strings.Split(data.Msg, "~")

				//
				// Parse the code in the message, which affects the
				// size of the split message that was received.
				//
				code, _ := strconv.Atoi(split[0])

				switch code {
				//
				// Zero (0) is a message with Trade data
				//
				case 0:
					bt.SendTrade(split)
					break

				//
				// Two (2) is a message with Current data
				//
				case 2:
					bt.SendCurrent(split)
					break

				//
				// Five (5) is a message with Current Aggregate data
				//
				case 5:
					bt.SendCurrentAggregate(split)
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

func (bt *Cryptocompbeat) SendTrade(data []string) {

	flag_int, _ := strconv.ParseInt(data[4], 10, 8)

	flag := "unknown"

	switch flag_int {
	case 1:
		flag = "buy"
		break

	case 2:
		flag = "sell"
		break
	}

	trade, _ := strconv.ParseInt(data[5], 10, 32)
	time_, _ := strconv.ParseInt(data[6], 10, 32)

	logp.Info(fmt.Sprintf(
		"TRADE %s -> %s %s (%s)",
		data[2],
		data[3],
		data[8],
		data[1],
	))

	fields := common.MapStr{
		"type":        "trade",
		"exchange":    data[1],
		"symbol_from": data[2],
		"symbol_to":   data[3],
		"flag":        flag,
		"trade":       trade,
		"time":        time_,
		"quantity":    data[7],
		"rate":        data[8],
		"total":       data[9],
	}

	event := beat.Event{
		Timestamp: time.Now(),
		Fields:    fields,
	}

	bt.client.Publish(event)

}

func (bt *Cryptocompbeat) SendCurrent(data []string) {
	bt.SendCurrentTemplate("current", data)
}

func (bt *Cryptocompbeat) SendCurrentAggregate(data []string) {
	bt.SendCurrentTemplate("current_aggregate", data)
}

func (bt *Cryptocompbeat) SendCurrentTemplate(type_ string, data []string) {

	flag_int, _ := strconv.ParseInt(data[4], 10, 8)

	flag := "unknown"

	switch flag_int {
	case 1:
		flag = "price_up"
		break

	case 2:
		flag = "price_down"
		break

	case 4:
		flag = "price_unchanged"
		break
	}

	last_update, _ := strconv.ParseInt(data[6], 10, 32)
	last_trade_id, _ := strconv.ParseInt(data[9], 10, 32)

	mask_int, _ := strconv.ParseInt(data[12], 10, 32)

	logp.Info(fmt.Sprintf(
		"CURRENT %s",
		type_,
	))

	fields := common.MapStr{
		"type":           type_,
		"exchange":       data[1],
		"symbol_from":    data[2],
		"symbol_to":      data[3],
		"flag":           flag,
		"rate":           data[5],
		"time":           last_update,
		"volume_from":    data[7],
		"volume_to":      data[8],
		"trade":          last_trade_id,
		"volume_24_from": data[10],
		"volume_24_to":   data[11],
		"mask_int":       mask_int,
	}

	event := beat.Event{
		Timestamp: time.Now(),
		Fields:    fields,
	}

	bt.client.Publish(event)
}

func (bt *Cryptocompbeat) Stop() {
	//bt.client.Close()
	close(bt.done)
}

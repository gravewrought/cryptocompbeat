package beater

import (
	"fmt"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Cryptocompmsg struct {
	Subs []string `json:"subs"`
}

type Cryptocompdata struct {
	Type int
	Msg  string
}

const (
	cryptocompare_endpoint = "streamer.cryptocompare.com"
	cryptocompare_port     = 443
	cryptocompare_ssl      = true
)

type Cryptocompsub struct {
	Type       string
	Exchange   string
	SymbolFrom string
	SymbolTo   string
	Data       chan Cryptocompdata
	Lock       chan bool
	Err        error
}

func (s *Cryptocompsub) Init() {

}

func (s *Cryptocompsub) Open() {
	s.Lock = make(chan bool)

	go s.dispatch()
}

func (s *Cryptocompsub) dispatch() {
	lock := make(chan bool)

	conn, err := gosocketio.
		Dial(gosocketio.GetUrl(cryptocompare_endpoint, cryptocompare_port, cryptocompare_ssl),
			transport.GetDefaultWebsocketTransport())

	s.Err = err

	if s.Err != nil {
		return
	}

	defer conn.Close()

	s.Err = conn.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		s.Data <- Cryptocompdata{Type: 0, Msg: "Connected"}

		s.Err = conn.
			Emit("SubAdd", Cryptocompmsg{Subs: []string{fmt.Sprintf("%s~%s~%s~%s", s.Type, s.Exchange, s.SymbolFrom, s.SymbolTo)}})
	})

	if s.Err != nil {
		return
	}

	s.Err = conn.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		s.Data <- Cryptocompdata{Type: 0, Msg: "Disconnected"}

		lock <- true
	})

	if s.Err != nil {
		return
	}

	s.Err = conn.On(gosocketio.OnError, func(h *gosocketio.Channel) {
		s.Data <- Cryptocompdata{Type: 0, Msg: "Error"}

		lock <- true
	})

	if s.Err != nil {
		return
	}

	s.Err = conn.On("m", func(c *gosocketio.Channel, msg string) {
		s.Data <- Cryptocompdata{Type: 1, Msg: msg}
	})

	if s.Err != nil {
		return
	}

	<-lock

	s.Lock <- true
}

func (s *Cryptocompsub) Close() {
}

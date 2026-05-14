package subscriber

import (
	"encoding/json"
	"log"

	"elibrary/pkg/eventbus"
)

type Subscriber struct {
	bus *eventbus.Publisher
}

func New(bus *eventbus.Publisher) *Subscriber {
	return &Subscriber{bus: bus}
}

func (s *Subscriber) Start() error {
	if s.bus == nil {
		return nil
	}

	return s.bus.Subscribe("user.created", func(_ string, data []byte) {
		var ev eventbus.UserCreatedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			log.Printf("[sub] bad user.created payload: %v", err)
			return
		}

		log.Printf("[sub] user.created received: user_id=%s email=%s", ev.UserID, ev.Email)
	})
}
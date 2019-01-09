package handler

import (
	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	l, _ := zap.NewProduction()
	logger = l.Sugar()
}

type NatsPublisher interface {
	Publish(subject string, v interface{}) error
}

type NatsSubscriber interface {
	Subscribe(subject string, cb nats.Handler) (*nats.Subscription, error)
}

type NatsPubSuber interface {
	NatsPublisher
	NatsSubscriber
}

// common pub/sub & log patterns
func subscribe(n NatsSubscriber, subject string, handler nats.Handler) {
	if _, err := n.Subscribe(subject, handler); err != nil {
		logger.Fatalf("failed to subscribe to subject=%s: %v", subject, err)
	}
	logger.Infof("Subscribed to subject=%s", subject)
}

func publishReply(n NatsPublisher, subject, reply string, resp interface{}) {
	if err := n.Publish(reply, resp); err != nil {
		logger.Errorf("could not publish reply to nats subject=%s reply=%s: %v", subject, reply, err)
	}
}

func publish(n NatsPublisher, subject string, event interface{}) {
	if err := n.Publish(subject, event); err != nil {
		logger.Errorf("could not publish to nats subject=%s: %v", subject, err)
	}
}
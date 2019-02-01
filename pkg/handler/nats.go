package handler

import (
	"github.com/nats-io/go-nats"
	"github.com/rs/zerolog/log"
)

type natsPublisher interface {
	Publish(subject string, v interface{}) error
}

type natsSubscriber interface {
	QueueSubscribe(subject, queue string, cb nats.Handler) (*nats.Subscription, error)
}

type natsPubSuber interface {
	natsPublisher
	natsSubscriber
}

// common pub/sub & log patterns
func queueSubscribe(n natsSubscriber, subject string, handler nats.Handler) {
	if _, err := n.QueueSubscribe(subject, "kube-nats", handler); err != nil {
		log.Fatal().Err(err).Str("subject", subject).Msg("could not subscribe to NATS subject")
	}
	log.Info().Str("subject", subject).Msg("subscribed to NATS subject")
}

func publishReply(n natsPublisher, subject, reply string, resp interface{}, respErr error) {
	if respErr != nil {
		resp = ErrorResp{respErr.Error()}
	}
	if err := n.Publish(reply, resp); err != nil {
		log.Error().
			Err(err).
			Str("subject", subject).
			Str("reply", reply).
			Msg("could not publish to NATS")
	}
}

type ErrorResp struct {
	Error string `json:"error"`
}

func publish(n natsPublisher, subject string, event interface{}) {
	if err := n.Publish(subject, event); err != nil {
		log.Error().
			Err(err).
			Str("subject", subject).
			Msg("could not publish to NATS")
	}
}

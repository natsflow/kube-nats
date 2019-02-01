package main

import (
	"github.com/nats-io/go-nats"
	"github.com/natsflow/kube-nats/pkg/handler"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"os"
	"strconv"
)

func main() {
	natsURL, ok := os.LookupEnv("NATS_URL")
	if !ok {
		natsURL = nats.DefaultURL
	}
	n := newNatsConn(natsURL)
	defer n.Close()

	cluster, ok := os.LookupEnv("CLUSTER")
	if !ok {
		log.Fatal().Msg("You must specify what kube cluster this is running by setting $CLUSTER")
	}

	k, err := newK8sCli()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create kubernetes client")
	}

	pubEvents, _ := strconv.ParseBool(os.Getenv("PUBLISH_EVENTS"))
	if pubEvents {
		go handler.WatchEvents(n, cluster, k)
	}

	go handler.Get(n, cluster, k)
	go handler.List(n, cluster, k)
	go handler.Create(n, cluster, k)
	go handler.Delete(n, cluster, k)

	select {}
}

func newNatsConn(url string) *nats.EncodedConn {
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("url", url).
			Msg("Failed to connect to NATS")
	}
	log.Info().Str("url", url).Msg("Connected to NATS")

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create NATS json connection")
	}
	return ec
}

func newK8sCli() (dynamic.Interface, error) {
	// TODO: allow running outside of cluster?
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

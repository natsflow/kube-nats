package main

import (
	nt "github.com/nats-io/go-nats"
	"github.com/natsflow/kube-nats/pkg/handler"
	"github.com/natsflow/kube-nats/pkg/k8s"
	"github.com/natsflow/kube-nats/pkg/nats"
	"go.uber.org/zap"
	"os"
)

var logger *zap.SugaredLogger

func init() {
	l, _ := zap.NewProduction()
	logger = l.Sugar()
}

func main() {
	natsURL, ok := os.LookupEnv("NATS_URL")
	if !ok {
		natsURL = nt.DefaultURL
	}
	n := nats.NewConnection(natsURL)
	defer n.Close()

	cluster, ok := os.LookupEnv("CLUSTER")
	if !ok {
		logger.Fatal("You must specify what cluster this is running by setting $CLUSTER")
	}

	k, err := k8s.NewDynamicClient()
	if err != nil {
		logger.Fatalf("Could not create kubernetes client: %s", err)
	}

	go handler.Get(n, cluster, k)
	go handler.List(n, cluster, k)
	go handler.Create(n, cluster, k)
	go handler.Delete(n, cluster, k)
	go handler.WatchEvents(n, cluster, k)

	select {}
}

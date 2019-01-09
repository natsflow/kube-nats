package main

import (
	"github.com/nats-io/go-nats"
	"github.com/natsflow/kube-nats/pkg/handler"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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
		natsURL = nats.DefaultURL
	}
	n := newNatsConn(natsURL)
	defer n.Close()

	cluster, ok := os.LookupEnv("CLUSTER")
	if !ok {
		logger.Fatal("You must specify what cluster this is running by setting $CLUSTER")
	}

	k, err := newK8sCli()
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

func newNatsConn(url string) *nats.EncodedConn {
	nc, err := nats.Connect(url)
	if err != nil {
		logger.Fatalf("Failed to connect to nats on %q: %v", url, err)
	}
	logger.Infof("Connected to nats %s", url)

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		logger.Fatalf("Failed to get json encoder: %v", err)
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

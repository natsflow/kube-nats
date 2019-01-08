package handler

import (
	nt "github.com/natsflow/kube-nats/pkg/nats"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type ErrorResp struct {
	Error string `json:"error"`
}

func List(n nt.PubSub, cluster string, k dynamic.Interface) {
	nt.Subscribe(n, "kube.list", listHandler(n, cluster, k))
}

func listHandler(n nt.Publisher, cluster string, k dynamic.Interface) func(subject, reply string, req ListReq) {
	return func(subject, reply string, req ListReq) {
		if req.Cluster != cluster {
			return
		}
		ul, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).List(req.ListOptions)
		if err != nil {
			nt.PublishReply(n, subject, reply, ErrorResp{err.Error()})
		}
		nt.PublishReply(n, subject, reply, ul)
	}
}

type ListReq struct {
	Cluster              string                      `json:"cluster"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
	Namespace            string                      `json:"namespace"`
	ListOptions          metav1.ListOptions          `json:"listOptions"`
}

func Get(n nt.PubSub, cluster string, k dynamic.Interface) {
	nt.Subscribe(n, "kube.get", getHandler(n, cluster, k))
}

func getHandler(n nt.Publisher, cluster string, k dynamic.Interface) func(subject, reply string, req GetReq) {
	return func(subject, reply string, req GetReq) {
		if req.Cluster != cluster {
			return
		}
		u, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Get(req.Name, req.GetOptions, req.Subresources...)
		if err != nil {
			nt.PublishReply(n, subject, reply, ErrorResp{err.Error()})
		}
		nt.PublishReply(n, subject, reply, u)
	}
}

type GetReq struct {
	Cluster              string                      `json:"cluster"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
	Namespace            string                      `json:"namespace"`
	Name                 string                      `json:"name"`
	GetOptions           metav1.GetOptions           `json:"getOptions"`
	Subresources         []string                    `json:"subresources"`
}

func Create(n nt.PubSub, cluster string, k dynamic.Interface) {
	nt.Subscribe(n, "kube.create", createHandler(n, cluster, k))
}

func createHandler(n nt.Publisher, cluster string, k dynamic.Interface) func(subject, reply string, req CreateReq) {
	return func(subject, reply string, req CreateReq) {
		if req.Cluster != cluster {
			return
		}
		u, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Create(req.Resource, req.CreateOptions, req.Subresources...)
		if err != nil {
			nt.PublishReply(n, subject, reply, ErrorResp{err.Error()})
		}
		nt.PublishReply(n, subject, reply, u)
	}
}

type CreateReq struct {
	Cluster              string                      `json:"cluster"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
	Namespace            string                      `json:"namespace"`
	Resource             *unstructured.Unstructured  `json:"resource"`
	CreateOptions        metav1.CreateOptions        `json:"createOptions"`
	Subresources         []string                    `json:"subresources"`
}

func Delete(n nt.PubSub, cluster string, k dynamic.Interface) {
	nt.Subscribe(n, "kube.delete", deleteHandler(n, cluster, k))
}

func deleteHandler(n nt.Publisher, cluster string, k dynamic.Interface) func(subject, reply string, req DeleteReq) {
	return func(subject, reply string, req DeleteReq) {
		if req.Cluster != cluster {
			return
		}
		err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Delete(req.Name, req.DeleteOptions, req.Subresources...)
		if err != nil {
			nt.PublishReply(n, subject, reply, ErrorResp{err.Error()})
		}
		nt.PublishReply(n, subject, reply, struct{}{})
	}
}

type DeleteReq struct {
	Cluster              string                      `json:"cluster"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
	Namespace            string                      `json:"namespace"`
	Name                 string                      `json:"name"`
	DeleteOptions        *metav1.DeleteOptions       `json:"deleteOptions"`
	Subresources         []string                    `json:"subresources"`
}

// publish's all kube events observed in the cluster
// i.e. what would be visible from `kubectl get events --all-namespaces -w`
// N.B. not all things that happen in kube get raised as events
// - see e.g. https://kubernetes.io/blog/2018/01/reporting-errors-using-kubernetes-events/ on how to raise
// N.B. as an aside; events have a `involvedObject:` that identifies the resource the event relate to
// - if you `kubectl describe` that resource you will see the specific events that relate to it
func WatchEvents(n nt.PubSub, cluster string, k dynamic.Interface) error {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "events",
	}
	lo := metav1.ListOptions{}
	watcher, err := k.Resource(gvr).Watch(lo)
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for e := range watcher.ResultChan() {
		resp := WatchEvent{
			Cluster: cluster,
			Event:   e,
		}
		nt.Publish(n, "kube.event.watch", resp)
	}
	return nil
}

type WatchEvent struct {
	Cluster string `json:"cluster"`
	watch.Event
}

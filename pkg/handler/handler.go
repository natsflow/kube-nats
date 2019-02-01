package handler

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

func List(n natsPubSuber, cluster string, k dynamic.Interface) {
	queueSubscribe(n, "kube.list", listHandler(n, cluster, k))
}

func listHandler(n natsPubSuber, cluster string, k dynamic.Interface) func(subject, reply string, req ListReq) {
	return func(subject, reply string, req ListReq) {
		if req.Cluster != cluster {
			return
		}
		ul, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).List(req.ListOptions)
		publishReply(n, subject, reply, ul, err)
	}
}

type ListReq struct {
	Cluster              string                      `json:"cluster"`
	GroupVersionResource schema.GroupVersionResource `json:"groupVersionResource"`
	Namespace            string                      `json:"namespace"`
	ListOptions          metav1.ListOptions          `json:"listOptions"`
}

func Get(n natsPubSuber, cluster string, k dynamic.Interface) {
	queueSubscribe(n, "kube.get", getHandler(n, cluster, k))
}

func getHandler(n natsPubSuber, cluster string, k dynamic.Interface) func(subject, reply string, req GetReq) {
	return func(subject, reply string, req GetReq) {
		if req.Cluster != cluster {
			return
		}
		u, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Get(req.Name, req.GetOptions, req.Subresources...)
		publishReply(n, subject, reply, u, err)
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

func Create(n natsPubSuber, cluster string, k dynamic.Interface) {
	queueSubscribe(n, "kube.create", createHandler(n, cluster, k))
}

func createHandler(n natsPubSuber, cluster string, k dynamic.Interface) func(subject, reply string, req CreateReq) {
	return func(subject, reply string, req CreateReq) {
		if req.Cluster != cluster {
			return
		}
		u, err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Create(req.Resource, req.CreateOptions, req.Subresources...)
		publishReply(n, subject, reply, u, err)
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

func Delete(n natsPubSuber, cluster string, k dynamic.Interface) {
	queueSubscribe(n, "kube.delete", deleteHandler(n, cluster, k))
}

func deleteHandler(n natsPubSuber, cluster string, k dynamic.Interface) func(subject, reply string, req DeleteReq) {
	return func(subject, reply string, req DeleteReq) {
		if req.Cluster != cluster {
			return
		}
		err := k.Resource(req.GroupVersionResource).Namespace(req.Namespace).Delete(req.Name, req.DeleteOptions, req.Subresources...)
		publishReply(n, subject, reply, struct{}{}, err)
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

// publishes all kube events observed in the cluster
// i.e. what would be visible from `kubectl get events --all-namespaces -w`
// N.B. not all things that happen in kube get raised as events
// - see e.g. https://kubernetes.io/blog/2018/01/reporting-errors-using-kubernetes-events/ on how to raise
// N.B. as an aside; events have a `involvedObject:` that identifies the resource the event relate to
// - if you `kubectl describe` that resource you will see the specific events that relate to it
func WatchEvents(n natsPublisher, cluster string, k dynamic.Interface) error {
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
		publish(n, "kube.event.watch", resp)
	}
	return nil
}

type WatchEvent struct {
	Cluster string `json:"cluster"`
	watch.Event
}

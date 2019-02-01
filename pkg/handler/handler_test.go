package handler

import (
	"errors"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"sync"
	"testing"
)

func TestListHandler(t *testing.T) {
	// given
	kubeResp := &unstructured.UnstructuredList{}
	r := new(ResourceInterfaceMock)
	r.On("List", metav1.ListOptions{}).Return(kubeResp, nil)
	n := new(NamespaceableResourceInterfaceMock)
	n.On("Namespace", "foo").Return(r)
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	i.On("Resource", gvr).Return(n)
	nts := new(NatsMock)
	nts.On("Publish", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", kubeResp).Return(nil)

	// when
	req := ListReq{
		Cluster:              "kube-cluster-1",
		Namespace:            "foo",
		GroupVersionResource: gvr,
	}
	listHandler(nts, "kube-cluster-1", i)("kube.list", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	nts.AssertExpectations(t)
}

func TestShouldOnlyHandleCurrentCluster(t *testing.T) {
	// given
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	nts := new(NatsMock)

	// when
	req := ListReq{
		Cluster:              "DIFFERENT-kube-cluster-",
		Namespace:            "foo",
		GroupVersionResource: gvr,
	}
	listHandler(nts, "kube-cluster-1", i)("kube.list", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertNotCalled(t, "Resource")
	nts.AssertNotCalled(t, "Publish")
}

func TestShouldPublishKubeErrorsToNats(t *testing.T) {
	// given
	err := errors.New("failed to list resource")
	r := new(ResourceInterfaceMock)
	r.On("List", metav1.ListOptions{}).Return(&unstructured.UnstructuredList{}, err) // returns error
	n := new(NamespaceableResourceInterfaceMock)
	n.On("Namespace", "foo").Return(r)
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	i.On("Resource", gvr).Return(n)
	nts := new(NatsMock)
	expectedErrResp := ErrorResp{
		err.Error(),
	}
	nts.On("Publish", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", expectedErrResp).Return(nil) //publishes error

	// when
	req := ListReq{
		Cluster:              "kube-cluster-1",
		Namespace:            "foo",
		GroupVersionResource: gvr,
	}
	listHandler(nts, "kube-cluster-1", i)("kube.list", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	nts.AssertExpectations(t)
}

func TestGetHandler(t *testing.T) {
	// given
	r := new(ResourceInterfaceMock)
	kubeResp := &unstructured.Unstructured{}
	r.On("Get", "nats-cluster-1", metav1.GetOptions{}, *new([]string)).Return(kubeResp, nil)
	n := new(NamespaceableResourceInterfaceMock)
	n.On("Namespace", "foo").Return(r)
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	i.On("Resource", gvr).Return(n)
	nts := new(NatsMock)
	nts.On("Publish", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", kubeResp).Return(nil)

	// when
	req := GetReq{
		Cluster:              "kube-cluster-1",
		Namespace:            "foo",
		GroupVersionResource: gvr,
		Name:                 "nats-cluster-1",
	}
	getHandler(nts, "kube-cluster-1", i)("kube.get", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	nts.AssertExpectations(t)
}

func TestCreateHandler(t *testing.T) {
	// given
	r := new(ResourceInterfaceMock)
	resourceToCreate := &unstructured.Unstructured{}
	kubeResp := &unstructured.Unstructured{}
	r.On("Create", resourceToCreate, metav1.CreateOptions{}, *new([]string)).Return(kubeResp, nil)

	n := new(NamespaceableResourceInterfaceMock)
	n.On("Namespace", "foo").Return(r)
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	i.On("Resource", gvr).Return(n)
	nts := new(NatsMock)
	nts.On("Publish", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", kubeResp).Return(nil)

	// when
	req := CreateReq{
		Cluster:              "kube-cluster-1",
		Namespace:            "foo",
		GroupVersionResource: gvr,
		Resource:             resourceToCreate,
	}
	createHandler(nts, "kube-cluster-1", i)("kube.create", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	nts.AssertExpectations(t)
}

func TestDeleteHandler(t *testing.T) {
	// given
	r := new(ResourceInterfaceMock)
	r.On("Delete", "nginx-5bd8487f5f-5spdh", &metav1.DeleteOptions{}, *new([]string)).Return(nil)

	n := new(NamespaceableResourceInterfaceMock)
	n.On("Namespace", "foo").Return(r)
	i := new(DynamicInterfaceMock)
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	i.On("Resource", gvr).Return(n)
	nts := new(NatsMock)
	nts.On("Publish", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", struct{}{}).Return(nil)

	// when
	req := DeleteReq{
		Cluster:              "kube-cluster-1",
		Namespace:            "foo",
		GroupVersionResource: gvr,
		Name:                 "nginx-5bd8487f5f-5spdh",
		DeleteOptions:        &metav1.DeleteOptions{},
	}
	deleteHandler(nts, "kube-cluster-1", i)("kube.delete", "_INBOX.aPXtkzW5ztAmDooWta7P1B.hXe5Q0m7", req)

	// then
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	nts.AssertExpectations(t)
}

func TestWatchEvents(t *testing.T) {
	// given
	kubeEventsChan := make(chan watch.Event)
	var kubeEventEmitChan <-chan watch.Event = kubeEventsChan // convert to send events only (required by method sig)
	w := new(WatchInterfaceMock)
	w.On("ResultChan").Return(kubeEventEmitChan)
	r := new(ResourceInterfaceMock)
	r.On("Watch", metav1.ListOptions{}).Return(w)
	n := new(NamespaceableResourceInterfaceMock)
	n.ResourceInterfaceMock = *r
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "events",
	}
	i := new(DynamicInterfaceMock)
	i.On("Resource", gvr).Return(n)

	// WatchEvents() (and therefore Publish()) is run in a goroutine, so we need to use a mutex to signal when
	// Publish() has actually been called
	mutex := &sync.Mutex{}
	mutex.Lock()
	nts := &natsPubStub{mutex: mutex}

	// when
	go WatchEvents(nts, "kube-cluster-1", i)
	ev := watch.Event{Type: watch.Added}
	kubeEventsChan <- ev //send some events

	// then
	// wait until Publish() has been called, otherwise actualSubject and actualResp will not have been set
	mutex.Lock()
	assert.Equal(t, "kube.event.watch", nts.actualSubject)
	assert.Equal(t, ev, nts.actualResp.(WatchEvent).Event)
	i.AssertExpectations(t)
	n.AssertExpectations(t)
	r.AssertExpectations(t)
	w.AssertExpectations(t)
}

type NatsMock struct {
	mock.Mock
}

func (n *NatsMock) Publish(subject string, v interface{}) error {
	args := n.Called(subject, v)
	return args.Error(0)
}

func (n *NatsMock) QueueSubscribe(subject, queue string, cb nats.Handler) (*nats.Subscription, error) {
	args := n.Called(subject, queue, cb)
	return args.Get(0).(*nats.Subscription), args.Error(1)
}

type DynamicInterfaceMock struct {
	mock.Mock
}

func (i *DynamicInterfaceMock) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	args := i.Called(resource)
	return args.Get(0).(dynamic.NamespaceableResourceInterface)
}

type NamespaceableResourceInterfaceMock struct {
	mock.Mock
	ResourceInterfaceMock
}

func (n *NamespaceableResourceInterfaceMock) Namespace(namespace string) dynamic.ResourceInterface {
	args := n.Called(namespace)
	return args.Get(0).(dynamic.ResourceInterface)
}

type ResourceInterfaceMock struct {
	mock.Mock
}

func (r *ResourceInterfaceMock) Create(obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := r.Called(obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (r *ResourceInterfaceMock) Update(obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := r.Called(obj, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (r *ResourceInterfaceMock) UpdateStatus(obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	args := r.Called(obj, options)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (r *ResourceInterfaceMock) Delete(name string, options *metav1.DeleteOptions, subresources ...string) error {
	args := r.Called(name, options, subresources)
	return args.Error(0)
}

func (r *ResourceInterfaceMock) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := r.Called(options, listOptions)
	return args.Error(0)
}

func (r *ResourceInterfaceMock) Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := r.Called(name, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (r *ResourceInterfaceMock) List(opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	args := r.Called(opts)
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}

func (r *ResourceInterfaceMock) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	args := r.Called(opts)
	return args.Get(0).(watch.Interface), nil
}

func (r *ResourceInterfaceMock) Patch(name string, pt types.PatchType, data []byte, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := r.Called(name, pt, data, options, subresources)
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

type WatchInterfaceMock struct {
	mock.Mock
}

func (w *WatchInterfaceMock) Stop() {
	w.Called()
}

func (w *WatchInterfaceMock) ResultChan() <-chan watch.Event {
	args := w.Called()
	return args.Get(0).(<-chan watch.Event)
}

type natsPubStub struct {
	actualSubject string
	actualResp    interface{}
	mutex         *sync.Mutex
}

func (n *natsPubStub) Publish(subject string, v interface{}) error {
	n.actualSubject = subject
	n.actualResp = v
	n.mutex.Unlock()
	return nil
}

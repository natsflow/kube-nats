# Kube Nats

Interact with kubernetes using NATS


## Quick start

Run NATS using e.g. [NATS Operator](https://github.com/nats-io/nats-operator) 
(this example assumes a NATS cluster running behind a service `nats-cluster`)

Run kube-nats:

```
skaffold dev
```

Interact with kube using nats. 
Here we use [node-nats](https://github.com/nats-io/node-nats) to log all kubernetes events from all namespaces in all clusters to the console:

```js
let NATS = require('nats')
let nats = NATS.connect({ 'json': true });
nats.subscribe('kube.event.watch', resp => console.log(resp))
```

By default kube-nats will connect to nats running on `nats://localhost:4222` - to change this set the `$NATS_URL`
env variable. 

Currently kube-nats is designed to run in and be responsible for a single cluster (it will use the `InClusterConfig` to connect to the kube api). 
You need to run a separate kube-nats instance for each cluster you have. As such, for disambiguation, the `$CLUSTER` env var needs to be defined for each
kubes-nats instance. This value is used in NATS messages to indicate from which cluster a message originated and also by 
kube-nats to know which requests it should handle. The local skaffold profile will use "minikube" as the value.

### Env Variables

key            | default value           | description
-------------- | ----------------------- | -----------
CLUSTER        | "minikube"              | The name of the kube cluster in which kube-nats is deployed
NATS_URL       | "nats://localhost:4222" | URL of the NATS server to connect to
PUBLISH_EVENTS | "false"                 | Whether kube-nats should publish kube watch events to NATS.  

If running multiple instances of kube-nats then you should set PUBLISH_EVENTS to "true" for at most one instance - this 
is to prevent duplicate events being published to NATS for a single Kubernetes event. The request-reply behaviour of kube-nats
uses [queue grouping](https://nats.io/documentation/concepts/nats-queueing/), hence natively handles running multiple instances.

## Nats Subjects

kube-nats uses the Kubernetes Go client's [dynamic kubernetes api](https://github.com/kubernetes/client-go/blob/master/dynamic/interface.go).
The types returned from kube-nats are the exact types that the library returns, serialised to json i.e. `*unstructured.Unstructured` or `*unstructured.UnstructuredList` 
These are the same as the responses you would receive if using the rest api directly or kubectl.

The following nats subjects are currently supported.

### Request-Reply

All message responses are json.

Note that the `groupVersionResource` object requires the *plural* name of a resource (e.g. 'pods', 'deployments' etc) - it will not
work with the singular versions.

For all subject requests, if an error occurs an object with a single string field "error" will be returned. e.g.

```
{
  "error": "deployments.apps \"nginx\" not found"
}
```

#### kube.list

List all kube resources matching the request. If you do not provide a namespace, then all namespaces will be searched. 
`groupVersionResource` is mandatory.
Response is essentially equivalent to e.g. `kubectl get pods -n default -o json`
For the exact request supported see [ListReq](pkg/handler/handler.go).
 
<details>
 <summary>e.g. (node)</summary>

```js
let req = {
  cluster: 'minikube',
  groupVersionResource: { Group: '', Version: 'v1', Resource: 'pods' },
  namespace: 'default',
  listOptions: {}
}
nats.requestOne('kube.list', req, {}, 3000, resp => {
  console.log(JSON.stringify(resp, null, 2))
})
```

output:

```
{
  "apiVersion": "v1",
  "items": [
    {
      "apiVersion": "v1",
      "kind": "Pod",
      "metadata": {
        "creationTimestamp": "2019-01-08T16:51:20Z",
        "generateName": "kube-nats-7c8599f6d5-",
        "labels": {
          "app": "kube-nats",
          "pod-template-hash": "7c8599f6d5"
        },
        "name": "kube-nats-7c8599f6d5-xf9cf",
        "namespace": "default",

         ...etc
```

</details> 
  
#### kube.get

Get the kube resource matching the request. If you do not provide a namespace, then all namespaces will be searched. 
`groupVersionResource` is mandatory.
Response is essentially equivalent to e.g. `kubectl get pod my-amazing-app-74c459d9d6-m828p -n foo -o json`
For the exact request supported see [GetReq](pkg/handler/handler.go).

<details>
 <summary>e.g. (node)</summary>

```js
let req = {
  cluster: 'minikube',
  groupVersionResource: { Group: '', Version: 'v1', Resource: 'pods' },
  namespace: 'default',
  name: 'nats-cluster-1'
}
nats.requestOne('kube.get', req, {}, 3000, resp => {
  console.log(JSON.stringify(resp, null, 2))
})
```

output:

```
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "annotations": {
      "nats.version": "1.3.0"
    },
    "creationTimestamp": "2019-01-08T16:49:40Z",
    "labels": {
      "app": "nats",
      "nats_cluster": "nats-cluster",
      "nats_version": "1.3.0"
    },
    "name": "nats-cluster-1",
    "namespace": "default",

         ...etc
```

</details> 

#### kube.create

Create the provided kube resource. 
`groupVersionResource` is mandatory.
Is essentially equivalent to e.g. `kubectl create deployment -f deploy.json`
For the exact request supported see [CreateReq](pkg/handler/handler.go).

<details>
 <summary>e.g. (node)</summary>

```js
let req = {
  cluster: 'minikube',
  groupVersionResource: { Group: 'apps', Version: 'v1', Resource: 'deployments' },
  namespace: 'default',
  resource: {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: {
      name: 'nginx',
      labels: { app: 'nginx' }
    },
    spec: {
      replicas: 1,
      selector: {
        matchLabels: { app: 'nginx' },
      },
      template: {
        metadata: {
          labels: { app: 'nginx' },
        },
        spec: {
          containers: [
            {
              name: 'nginx',
              image: 'nginx:latest'
            }
          ]
        }
      }
    }
  },
  createOptions: {},
  subresources: [],
}
nats.requestOne('kube.create', req, {}, 3000, resp => {
  console.log(JSON.stringify(resp, null, 2))
})
```

output:

```
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "creationTimestamp": "2019-01-08T17:12:22Z",
    "generation": 1,
    "labels": {
      "app": "nginx"
    },
    "name": "nginx",
    "namespace": "default",
    "resourceVersion": "3691",
         ...etc
```

</details> 

#### kube.delete

Delete the specified kube resource. 
`groupVersionResource` is mandatory.
Is essentially equivalent to e.g. `kubectl delete deploy my-amazing-app -n foo`
For the exact request supported see [DeleteReq](pkg/handler/handler.go).

<details>
 <summary>e.g. (node)</summary>

```js
req = {
  cluster: 'minikube',
  groupVersionResource: { Group: 'apps', Version: 'v1', Resource: 'deployments' },
  deleteOptions: { propagationPolicy: 'Foreground' },
  namespace: 'default',
  name: 'nginx',
}
nats.requestOne('kube.delete', req, {}, 3000, resp => {
  console.log(JSON.stringify(resp, null, 2))
})
```

output:

```
{}
```

</details> 

### Publish-Subscribe 

#### kube.event.watch

Watches all kubernetes events in all namespaces and all clusters (clusters must be running an instance of kube-nats).  
Response is essentially equivalent to `kubectl get events --all-namespaces -w -o json` (assuming that kubectl connected to all clusters!)
For the exact event published see [WatchEvent](pkg/handler/handler.go).

<details>
 <summary>e.g. (node)</summary>

```js
nats.subscribe('kube.event.watch', resp => {
  console.log(resp)
})
```

output:

```
{ cluster: 'minikube',
  Type: 'ADDED',
  Object:
   { apiVersion: 'v1',
     count: 1,
     eventTime: null,
     firstTimestamp: '2018-11-28T12:30:10Z',
     involvedObject:
      { apiVersion: 'v1',
        fieldPath: 'spec.containers{nats}',
        kind: 'Pod',
        name: 'nats-cluster-2',
        namespace: 'default',
        resourceVersion: '840',
        uid: 'b1e99aa5-f305-11e8-b09e-827816e0a801' },
     kind: 'Event',
     lastTimestamp: '2018-11-28T12:30:10Z',
     message: 'Killing container with id docker://nats:Need to kill Pod',
     ...
```

</details> 

# Legal
This project is available under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html).

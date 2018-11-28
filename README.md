# Kube Nats

Interact with kubernetes using NATS


## Quick start

Run NATS using e.g. [NATS Operator](https://github.com/nats-io/nats-operator) 
(this example assumes a NATS cluster running behind a service `example-nats-cluster`)

Run kube-nats:

```
skaffold dev
```

Interact with kube using nats. 
Here we log all kubernetes events from all namespaces in all clusters to the console:

```javascript
nats.subscribe('kube.event.watch', resp => console.log(resp))
```

By default kube-nats will connect to nats running on `nats://localhost:4222` - to change this set the `$NATS_URL`
env variable. 

Currently kube-nats is designed to run in and be responsible for a single cluster (it will use the `InClusterConfig` to connect to the kube api). 
You need to run a separate kube-nats instance for each cluster you have. As such, for disambiguation, the `$CLUSTER` env var needs to be defined for each
kubes-nats instance. This value is used in NATS messages to indicate from which cluster a message originated and also by 
kube-nats to know which requests it should handle. The local skaffold profile will use "minikube" as the value.

## Nats Subjects

kube-nats effectively proxies the [dynamic kubernetes api](https://github.com/kubernetes/client-go/blob/master/dynamic/interface.go)
The message bodies for requests & responses follow the corresponding kube api message bodies as closely as possible 

The following nats subjects are currently supported.

### Request-Reply

All message responses are json and will include a non-empty `err` string field in the case of an error.

#### kube.list

List all kube resources matching the request. If you do not provide a namespace, then all namespaces will be searched. 
`groupVersionResource` is mandatory.
Response is essentially equivalent to e.g. `kubectl get pods -n default -o json`
For the exact requests & responses supported see [ListReq & ListResp](pkg/handler/handler.go).
 
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
  "resources": {
    "apiVersion": "v1",
    "items": [
      {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
          "annotations": {
            "nats.version": "1.3.0"
          },
          "creationTimestamp": "2018-11-26T09:54:38Z",
          "labels": {
            "app": "nats",
            "nats_cluster": "example-nats-cluster",
            "nats_version": "1.3.0"
          },
          "name": "example-nats-cluster-1",
          "namespace": "default",
         ...etc
```

</details> 
  
#### kube.get

Get the kube resource matching the request. If you do not provide a namespace, then all namespaces will be searched. 
`groupVersionResource` is mandatory.
Response is essentially equivalent to e.g. `kubectl get pod my-amazing-app-74c459d9d6-m828p -n foo -o json`
For the exact requests & responses supported see [GetReq & GetResp](pkg/handler/handler.go).

<details>
 <summary>e.g. (node)</summary>

```js
let req = {
  cluster: 'minikube',
  groupVersionResource: { Group: '', Version: 'v1', Resource: 'pods' },
  namespace: 'default',
  name: 'example-nats-cluster-1'
}
nats.requestOne('kube.get', req, {}, 3000, resp => {
  console.log(JSON.stringify(resp, null, 2))
})
```

output:

```
{
  "resource": {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
      "annotations": {
        "nats.version": "1.3.0"
      },
      "creationTimestamp": "2018-11-26T09:54:38Z",
      "labels": {
        "app": "nats",
        "nats_cluster": "example-nats-cluster",
        "nats_version": "1.3.0"
      },
      "name": "example-nats-cluster-1",
      "namespace": "default",
         ...etc
```

</details> 

#### kube.create

Create the provided kube resource. 
`groupVersionResource` is mandatory.
Is essentially equivalent to e.g. `kubectl create deployment -f deploy.json`
For the exact requests & responses supported see [CreateReq & CreateResp](pkg/handler/handler.go).

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
  "resource": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "creationTimestamp": "2018-11-27T00:21:29Z",
      "generation": 1,
      "labels": {
        "app": "nginx"
      },
      "name": "nginx",
      "namespace": "default",
      "resourceVersion": "109926",
         ...etc
```

</details> 

#### kube.delete

Delete the specified kube resource. 
`groupVersionResource` is mandatory.
Is essentially equivalent to e.g. `kubectl delete deploy my-amazing-app -n foo`
For the exact requests & responses supported see [DeleteReq & DeleteResp](pkg/handler/handler.go).

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
{
  "err": ""
}

```

</details> 

### Publish-Subscribe 

#### kube.event.watch

Watches all kubernetes events in all namespaces and all clusters (clusters must be running an instance of kube-nats).  
Response is essentially equivalent to `kubectl get events --all-namespaces -w -o json` (assuming that kubectl connected to all clusters!)
For the exact response see [WatchEventResp](pkg/handler/handler.go).

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
        name: 'example-nats-cluster-2',
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

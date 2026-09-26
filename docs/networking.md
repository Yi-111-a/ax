# Networking

Tasks do not get a Kubernetes Service or Ingress of their own. Every request to a task goes through Agent Substrate's **atenet router**, the `atenet-router` Service in the `ate-system` namespace. The router requires the actor's stable DNS authority (`Host` for HTTP or `:authority` for gRPC) and accepts `ate-target-actor` as the explicit target identifier; it resumes the actor first if it was suspended, then proxies the request there.

The actor authority is `<task>.<atespace>.actors.resources.substrate.ate.dev`, and the target header is `<atespace>/<task>`. The controller always names a task's actor after the task, so both values below identify the task `task123` in the `default` atespace.

## From inside the cluster

Use the Service DNS name and set the actor authority. This is exactly how the controller polls a task's readiness.

```bash
curl -H "Host: task123.default.actors.resources.substrate.ate.dev" \
  -H "ate-target-actor: default/task123" \
  http://atenet-router.ate-system.svc.cluster.local/metadata/v1alpha1/ax/task
```

## From your machine

Port-forward the router, then talk to it the same way.

```bash
kubectl -n ate-system port-forward svc/atenet-router 8001:80
curl -H "Host: task123.default.actors.resources.substrate.ate.dev" \
  -H "ate-target-actor: default/task123" http://localhost:8001/readyz
```

## gRPC request routing

Send the header as outgoing metadata under the lowercase key. This is what `ax ssh` does to reach the guest services.

```go
ctx = metadata.AppendToOutgoingContext(ctx, "ate-target-actor", "default/task123")
resp, err := client.SomeMethod(ctx, req)
```

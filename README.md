# k8s-inventory-client

[![ci](https://github.com/neticdk-k8s/k8s-inventory-client/actions/workflows/main.yml/badge.svg)](https://github.com/neticdk-k8s/k8s-inventory-client/actions/workflows/main.yml)
[![tag](https://img.shields.io/github/tag/neticdk-k8s/k8s-inventory-client.svg)](https://github.com/neticdk-k8s/k8s-inventory-client/tags/)

Kubernetes application that collects, exposes and optionally uploads inventory
data.

It runs in a loop every `COLLECT_INTERNAL` and collects:

- Cluster information (versions, etc)
- Workload information (deployments, etc)
- Infrastructure information (cloud, region, zone, etc)
- Node information
- Storage information
- Secure Cloud Stack specific information (customer data, etc)
- Custom Resource information (CNI, specific operators)

For more information about what is collected, see
[k8s-inventory](https://github.com/neticdk-k8s/k8s-inventory).

## Running

Configure the client (see Configuration below) and run the executable. It
doesn't take any command line options or arguments:

```bash
k8s-inventory-client
```

k8s-inventory-client will try to detect if it's running inside a kubernetes
cluster. If not, it will try to configure a kubernetes client by using the
credentials and default context from `~/.kube/config`. This makes it possible to
run the client outside of a cluster, e.g. for development purposes.

## Pre-requisites

### Secure Cloud Stack Information

While not strictly required, not gathering Secure Cloud Stack information kind of defeats the purpose of running the client at all.

Secure Cloud Stack Information is read from `ConfigMap`-objects.

These objects must be created in the `netic-metadata-system` `Namespace`:

- General cluster information uses the first `ConfigMap` matching the `netic.dk/owned-by=operator` label
- Tenant information uses all `ConfigMap`s matching the `netic.dk/owned-by=tenant` label

See [`dist/deployment/netic-metadata.yaml`](dist/deployment/netic-metadata.yaml) for examples.

### Permissions

In order to collect the data, k8s-inventory-client needs permission to read
a wide array of resources. See [`dist/deployment/k8s-inventory-client.yaml`](dist/deployment/k8s-inventory-client.yaml)
for a list of permissions.

## Configuration

Configuration is done using environment variables:

| Variable Name         | Description                                      |                                Default |
| :-------------------- | :----------------------------------------------- | -------------------------------------: |
| `HTTP_PORT`           | HTTP port to listen on for the inventory service |                                   8087 |
| `HTTP_PORT_META`      | HTTP port to listen on for the metadata service  |                                   8088 |
| `COLLECT_INTERNAL`    | How often to collect                             |                                     1h |
| `LOG_LEVEL`           | Logging level                                    |                                   info |
| `LOG_FORMATTER`       | Log output formatter                             |                                   json |
| `UPLOAD_INVENTORY`    | Upload inventory                                 |                                   true |
| `SERVER_API_ENDPOINT` | HTTP URL to upload data to                       | http://localhost:8086/api/v1/inventory |
| `AUTH_ENABLED`        | Enable/disable authentication                    |                                   true |
| `TLS_CRT`             | PEM Certificate file to use for authentication   |              /etc/certificates/tls.crt |
| `TLS_KEY`             | PEM KEY file to use for authentication           |              /etc/certificates/tls.key |
| `SERVER_API_ENDPOINT` | HTTP URL to upload data to                       | http://localhost:8086/api/v1/inventory |
| `IMPERSONATE`         | Kubernetes role to imporsonate                   |                                        |

### Collection Intervals

`COLLECT_INTERNAL` takes a values that can be parsed by
[`time.ParseDuration()`](https://pkg.go.dev/time#Duration).

### Log Formatter

`LOG_FORMATTER` can be set to one of:

- json
- text

### Log Level

`LOG_LEVEL` can be set to one of:

- debug
- info
- warn
- error
- fatal
- panic

## Deploying

If you use a Secure Cloud Stack cluster, chances are k8s-inventory-client is already deployed.

However, if you need to deploy it yourself, you will need to:

- Create the `ConfigMap`s mentioned above
- Create the Kubernetes resources:
  - ServiceAccount
  - ClusterRole
  - ClusterRoleBinding
  - Service
  - Deployment

See [`kustomization.yaml`](dist/deployment/kustomization.yaml) for a full example of deploying the client.

[`netic-metadata.yaml`](dist/deployment/netic-metadata.yaml) creates the `ConfigMap`s.

[`k8s-inventory-client.yaml`](dist/deployment/k8s-inventory-client.yaml) creates the other resources.

Typical adjustments:

- Set the `app.kubernetes.io/version` label to match the image version
- Set the image tag to the version you want to deploy
- In the `Deployment` set environment variables accordingly
- If using another port, set the `Service` ports

## Building

Use the supplied `Makefile`.

`make build` or `make build2` builds the executable and places it in
`bin/k8s-inventory-client`. `build` forces rebuilding of packages and also runs
the linter. `build2` does not.

`make docker-build` and `make docker-push` builds/tags/pushes the docker image.

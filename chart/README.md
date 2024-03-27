# k8s-inventory-client

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.6.2](https://img.shields.io/badge/AppVersion-0.6.2-informational?style=flat-square)

Inventory Client Helm Chart

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Netic A/S | <support@netic.dk> |  |

## Source Code

* <https://github.com/neticdk-k8s/k8s-inventory-client>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| authEnabled | string | `"true"` | Whetherauthentication towards the inventory server is enabled. Enabling authentication also creates mounts the certificate volume. |
| clusterIDConfigMap | object | `{"clusterFQDN":"prod999.myprovider.k8s.netic.dk","clusterName":"prod999","providerName":"myprovider"}` | The keys and values used in the cluster-id Config |
| collectInterval | string | `"30m"` | How often to collect the inventory |
| createClusterIDConfigMap | bool | `false` | Whether to create the cluster-id ConfigMap |
| debugEnabled | string | `"false"` | Whether debug mode is enabled |
| enablePriorityClass | bool | `false` | Enables PriorityClass for the pod. |
| fullnameOverride | string | `""` |  |
| httpPort | int | `8087` | The port to listen on for dumping the current inventory |
| httpPortMeta | int | `8088` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"ghcr.io/neticdk-k8s/k8s-inventory-client"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| livenessProbe.httpGet.path | string | `"/"` |  |
| livenessProbe.httpGet.port | string | `"meta"` |  |
| logFormatter | string | `"json"` | The logging formatter ("json" or "text") |
| logLevel | string | `"info"` | The logging level |
| nameOverride | string | `""` |  |
| priorityClassName | string | `"secure-cloud-stack-technical-management-critical"` | The name of the PriorityClass to use. |
| readinessProbe.httpGet.path | string | `"/"` |  |
| readinessProbe.httpGet.port | string | `"meta"` |  |
| resources.limits.cpu | string | `"50m"` |  |
| resources.limits.memory | string | `"256Mi"` |  |
| resources.requests.cpu | string | `"10m"` |  |
| resources.requests.memory | string | `"160Mi"` |  |
| serverAPIEndPoint | string | `"http://localhost:8086"` | Where the inventory API can be found |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"k8s-inventory-client"` |  |
| uploadInventory | string | `"true"` | Whether the inventory should be uploaded |
| volumeMounts | list | `[{"mountPath":"/etc/certificates","name":"certificate","readOnly":true}]` | Volumes to mount. The certificate is required for authentication. |
| volumes[0].name | string | `"certificate"` |  |
| volumes[0].secret.secretName | string | `"k8s-inventory-client-certificate"` |  |


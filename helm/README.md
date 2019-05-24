### Installing the Chart

To install the chart with the release name `grafana` in namespace `monitoring`:

```console
$ helm upgrade grafana charts/grafana --namespace monitoring --install
```
The command deploys grafana and grafana-config-controller on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

### Uninstalling the Chart

To uninstall/delete the `grafana` deployment:

```console
$ helm delete grafana --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
The following table lists the configurable parameters of the grafana-config-controller chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of pod replicas | `1`
`global` | Set additional global values | `{}`
`global.namespace` | Overwrite release namespace | `{}`
`grafana.image.repository` | grafana container image repository | `grafana/grafana`
`grafana.image.tag` | grafana container image tag | `6.1.4`
`grafana.resources.limits.cpu` | grafana container resources limits for cpu | `2000m`
`grafana.resources.limits.memory` | grafana container resources limits for memory | `8Gi`
`grafana.resources.requests.cpu` | grafana container resources limits for cpu | `1000m`
`grafana.resources.requests.memory` | grafana container resources limits for memory | `4Gi`
`extraEnv` | extra environment variable list for grafana | `[]`
`grafanaController.image.repository` | grafana-controller container image repository | `dockerregistry/grafana-controller`
`grafanaController.image.tag` | grafana-controller container image tag | `1.1.0`
`grafanaController.url` | The internal url to access grafana | `http://localhost:3000`
`grafanaController.id` | The id to specify grafana | `0`
`grafanaController.logLevel` | The log-level of grafana-controller | `info`
`volumeClaimTemplates.name` | The name of Persistent Volume für Granfana storage | `data`
`volumeClaimTemplates.accessModes` | Granfana server data Persistent Volume access modes | `[ "ReadWriteOnce" ]`
`volumeClaimTemplates.requests.storage` | Granfana server data Persistent Volume size | `10Gi`
`ingress.enabled` | If true, create ingress | `false`
`ingress.url` | Grafana ingress host url | ``
`ingress.extraAnnotations` | Extra annotations for ingress | `{}`
`service.port` | The port the grafana uses | `3000`
`service.type` | Specify the service type | `ClusterIP`
`service.path` | The metrics path grafana uses | `/metrics`
`service.scrape` | If true, prometheus scrapes grafana metrics | `true`
`securityContext.fsGroup` | Custom security context for grafana container | `0`
`securityContext.runAsUser` | Custom user for grafana container | `0`
`terminationGracePeriodSeconds` | Grafana Pod termination grace period | `10`
`customconfig.grafanaini` | Custom configurated grafana.ini | `<grafana.ini>`
`adminPassword`| Specify password for user: admin | `password`
`monitoringPassword`| Specify password for user: monitoring (which has only read access) | `password`



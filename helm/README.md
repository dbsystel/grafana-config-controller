### Installing the Chart

To install the chart with the release name `grafanai-config-controller` in namespace `monitoring`:

```console
$ helm upgrade grafana-config-controller examples/helm/charts/grafana-config-controller --namespace monitoring --install
```
The command deploys grafana-config-controller on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

### Uninstalling the Chart

To uninstall/delete the `grafana-config-controller` deployment:

```console
$ helm delete grafana-config-controller --purge
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
The following table lists the configurable parameters of the grafana-config-controller chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`replicaCount` | The number of pod replicas | `1`
`image.repository` | grafana-config-controller container image repository | `dockerregistry/devops/grafana-config-controller`
`image.tag` | grafana-config-controller container image tag | `1.1.0`
`url` | The internal url to access grafana | `http://grafana:3000`
`id` | The id to specify grafana | `0`
`logLevel` | The log-level of grafana-config-controller | `info`
`adminPassword` | The admin password the grafana-config-controller uses to access Grafana apis | `adminpassword`

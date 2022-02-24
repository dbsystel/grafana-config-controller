:warning: The project has been archived and is no longer maintained!

# Config Controller for Grafana

This Controller is based on the [Grafana Operator](https://github.com/tsloughter/grafana-operator). 
The Config Controller should be run within [Kubernetes](https://github.com/kubernetes/kubernetes) as a sidecar with [Grafana](https://github.com/grafana/grafana).

It watches for new/updated/deleted *ConfigMaps* and if they define the specified annotations as `true` it will `POST` each resource from ConfigMap to Grafana API. This requires Grafana > 5.x.

## Annotations

Currently it support three resources:


**1. Dashboard**

`grafana.net/dashboard` with values: `"true"` or `"false"`

`grafana.net/folder` with values: `"true"`, `"false"` or `"customName"`:

`grafana.net/folder: "true"` = Dashboard will be loaded into a folder. Name of the folder is based on the Namespace in which the ConfigMap was loaded into or

`grafana.net/folder: "false"` = Dashboard will be loaded into the default location/folder named *General* or

`grafana.net/folder: "customName"` = Dashboard will be loaded into a folder. Name of the folder is based on provided `customName`


**2. Datasource**

`grafana.net/datasource` with values: `"true"` or `"false"`

**3. Notification Channel**

`grafana.net/notification-channel` with values: `"true"` or `"false"`

(**Id**)

`grafana.net/id` with values: `"0"` ... `"n"`

In case of multiple Grafana *setups* in same Kubernetes Cluster all the ConfigMaps have to be mapped to the right Grafana setup.
So each *ConfigMap* can be additionaly annotated with the `grafana.net/id` (if not, the default `id` will be `"0"`)

**Note**

Mentioned `"true"` values can be also specified with: `"1", "t", "T", "true", "TRUE", "True"`

Mentioned `"false"` values can be also specified with: `"0", "f", "F", "false", "FALSE", "False"`

**ConfigMap examples can be found [here](configmap-examples).**

## Usage
```
--log-level # desired log level, one of: [debug, info, warn, error]
--log-format # desired log format, one of: [json, logfmt]
--run-outside-cluster # Uses ~/.kube/config rather than in cluster configuration
--grafana-url # Sets the URL and authentication to use to access the Grafana API
--id # Sets the ID, so the Controller knows which ConfigMaps should be watched
--namespace # Only watch specified namespace
```

## Development
### Build
```
go build -v -i -o ./bin/grafana-config-controller ./cmd # on Linux
GOOS=linux CGO_ENABLED=0 go build -v -i -o ./bin/grafana-config-controller ./cmd # on macOS/Windows
```
To build a docker image out of it, look at provided [Dockerfile](Dockerfile) example.

## Deployment
Our preferred way to install grafana-config-controller is [Helm](https://helm.sh/). See example installation at our [Helm directory](helm) within this repo.

## Scripts
If you want to export grafana dashboards and datasources into json files you can use the provided [scripts](scripts) within this repo.


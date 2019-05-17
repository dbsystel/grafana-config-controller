FROM alpine:latest

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator

USER kube-operator

COPY ./bin/grafana-config-controller .

ENTRYPOINT ["./grafana-config-controller"]

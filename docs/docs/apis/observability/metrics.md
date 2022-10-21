---
title: Metrics
---

ZITADEL provides a `metrics` endpoint with the help of the [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) package.

Self-hosting customers can access this endpoint with on the path `/debug/metrics`. For example when running ZITADEL locally its is accessible on `http://localhost:8080/debug/metrics`. The metrics endpoint can be scrubbed by any tool of choice that supports the `otel` format, e.g  an existing Prometheus.

For our [Kubernetes/Helm](../../guides/deploy/kubernetes) users, we provide an out of the box support for the [ServiceMonitor](https://github.com/dennigogo/zitadel-charts/blob/main/charts/zitadel/templates/servicemonitor.yaml) custom resource.

By default, metrics are enabled but can be turned of through ZITADEL's [configuration](../../guides/manage/self-hosted/configure). The (default) configuration is located in the [defaults.yaml](https://github.com/dennigogo/zitadel/blob/main/cmd/defaults.yaml).

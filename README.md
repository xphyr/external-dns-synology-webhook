# external-dns-synology-webhook

**WARNING, WARNING, WARNING!**
**THIS WEBHOOK IS NOT READY FOR CONSUMPTION YET**

This project is webhook provider for the [Kubernetes external dns](https://github.com/kubernetes-sigs/external-dns) webhook provider for synology dns services. This project aims to create a webhook provider that will leverage the (undocumented) Synology API to add/edit/update/delete DNS records on the Synology Arrays using the "DNS" Service.

The project will build on the WebApi package that is a part of my fork of the [Synology CSI](https://github.com/xphyr/synology-csi) project. Synology is not currently maintaining this package so I have made a fork of it to keep it updated. I am also adding additional functionality to the `dsm\webapi` package in that repo. 

Note: If you are looking to get something running right now, see this blog post: https://blog.differentpla.net/blog/2025/05/03/k8s-external-dns-synology/

### Creating Synology Secret

Create a file called `external-dns-synology-secret.yaml` and update with the proper information:

``` yaml
---
apiVersion: v1
kind: Secret
metadata:
    name: external-dns-synology-secret
    namespace: external-dns
stringData:
  username: <your-username>
  password: <your-password>
```

### Installing the provider

1. Add the ExternalDNS Helm repository to your cluster.

    ```sh
    oc new-project external-dns
    helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
    oc apply -f external-dns-synology-secret.yaml
    ```

2. Deploy your `external-dns-synology-secret` secret that holds your authentication credentials from either of the credential types above.

3. Create the helm values file, for example `external-dns-synology-values.yaml`:

```yaml
fullnameOverride: external-dns-synology
logLevel: &logLevel debug
provider:
    name: webhook
    webhook:
    image:
        repository: ghcr.io/xphyr/external-dns-synology-webhook
        tag: main # replace with a versioned release tag
    env:
        - name: SYNOLOGY_HOSTNAME
        value: 192.168.1.1 # replace with the address to your Synology router/controller
        - name: SYNOLOGY_USERNAME
        valueFrom:
            secretKeyRef:
            name: external-dns-synology-secret
            key: username
        - name: SYNOLOGY_PASSWORD
        valueFrom:
            secretKeyRef:
            name: external-dns-synology-secret
            key: password
        - name: LOG_LEVEL
        value: *logLevel
    livenessProbe:
        httpGet:
        path: /healthz
        port: http-webhook
        initialDelaySeconds: 10
        timeoutSeconds: 5
    readinessProbe:
        httpGet:
        path: /readyz
        port: http-webhook
        initialDelaySeconds: 10
        timeoutSeconds: 5
extraArgs:
    - --ignore-ingress-tls-spec
policy: sync
sources: ["ingress", "service"]
txtOwnerId: default
txtPrefix: k8s.
domainFilters: ["example.com"] # replace with your domain(s), comma delimited
podSecurityContext:
  runAsNonRoot: true
  seccompProfile:
    type: RuntimeDefault
```

4. Install the Helm chart

    ```sh
    helm install external-dns-synology external-dns/external-dns -f external-dns-synology-values.yaml --version 1.19.0 -n external-dns
    ```

## Environment variables

The following environment variables are available:

| Variable           | Description                      | Notes                |
| ------------------ | -------------------------------- | -------------------- |
| SYNOLOGY_HOST_NAME | Synology Host Name or IP Address | Mandatory            |
| SYNOLOGY_USERNAME  | Synology Username                | Mandatory            |
| SYNOLOGY_PASSWORD  | Synology Password                | Mandatory            |
| SYNOLOGY_PORT      | Synology API Port                | Default: `5001`      |
| DRY_RUN            | If set, changes won't be applied | Default: `false`     |
| WEBHOOK_HOST       | Webhook hostname or IP address   | Default: `localhost` |
| WEBHOOK_PORT       | Webhook port                     | Default: `8888`      |
| HEALTH_HOST        | Liveness and readiness hostname  | Default: `0.0.0.0`   |
| HEALTH_PORT        | Liveness and readiness port      | Default: `8080`      |
| READ_TIMEOUT       | Servers' read timeout in ms      | Default: `60000`     |
| WRITE_TIMEOUT      | Servers' write timeout in ms     | Default: `60000`     |
| DOMAIN_FILTER      | List of acceptable domains       | Mandatory List       |

The Synology DNS Webhook will ONLY work on the zones/domains listed in the *DOMAIN_FILTER* env variable. You must specify a list of zones to work with. The Zones MUST match the "Zone ID" as listed in the Synology DNS Manager.


## Building & Manually Installing

### Using GoReleaser

This project uses GoReleaser to build the multi-arch container files.
To test: REPO_OWNER=<usernamehere> goreleaser release --snapshot --clean

## Inspiration

The following external dns webhook projects acted as inspiration for how to approach this project. Thanks to them for the inspiration on how to lay this project out.

* https://github.com/kashalls/external-dns-unifi-webhook
* https://github.com/mirceanton/external-dns-provider-mikrotik
* https://github.com/vultr/external-dns-vultr-webhook



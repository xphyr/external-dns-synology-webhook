# external-dns-synology-webhook

**WARNING, WARNING, WARNING!**
**THIS WEBHOOK IS NOT READY FOR CONSUMPTION YET**

This project is webhook provider for the [Kubernetes external dns](https://github.com/kubernetes-sigs/external-dns) webhook provider for synology dns services. This project aims to create a webhook provider that will leverage the (undocumented) Synology API to add/edit/update/delete DNS records on the Synology Arrays using the "DNS" Service.

The project will build on the WebApi package that is a part of my fork of the [Synology CSI](https://github.com/xphyr/synology-csi) project. Synology is not currently maintaining this package so I have made a fork of it to keep it updated. I am also adding additional functionality to the `dsm\webapi` package in that repo. 

Note: If you are looking to get something running right now, see this blog post: https://blog.differentpla.net/blog/2025/05/03/k8s-external-dns-synology/

## Building & Manually Installing

### Using GoReleaser

This project uses GoReleaser to build the multi-arch container files.

To test: REPO_OWNER=<usernamehere> goreleaser release --snapshot --clean

## Inspiration

The following external dns webhook projects acted as inspiration for how to approach this project. Thanks to them for the inspiration on how to lay this project out.

* https://github.com/kashalls/external-dns-unifi-webhook
* https://github.com/mirceanton/external-dns-provider-mikrotik
* https://github.com/vultr/external-dns-vultr-webhook



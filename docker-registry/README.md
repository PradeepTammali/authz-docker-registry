# Docker registry with certs 

You can update the existing configmap with your certs or you remove the configmap from yaml and create as using following command.

`kubectl -n <namespace> create configmap registry-certs --from-file=server.crt=server.crt --from-file=server.key=server.key`

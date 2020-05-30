#!/bin/bash

set -e

usage() {
    cat <<EOF
Generate certificate suitable for use with an sidecar-injector webhook service.

This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with sidecar-injector webhook
services. This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explantion and additional instructions.

The server key/cert k8s CA cert are stored in a k8s secret.

usage: ${0} [OPTIONS]

The following flags are required.

       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
       --certSecret       Secret name for CA certificate.
       --keySecret        Secret name for server certificate.
       --delete           Deletes the files and which were created.
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        --certSecret)
            certSecret="$2"
            shift
            ;;
        --keySecret)
            keySecret="$2"
            shift
            ;;
        --delete)
            delete="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done


[ -z ${service} ] && service=webhook
[ -z ${certSecret} ] && secret=cert
[ -z ${keySecret} ] && secret=key
[ -z ${namespace} ] && namespace=webhook

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

csrName=${service}.${namespace}
masterIP=$(kubectl get nodes --selector=node-role.kubernetes.io/master= -o jsonpath={.items[*].status.addresses[?\(@.type==\"InternalIP\"\)].address})

if [[ ${delete} == 'true' ]]; then
    echo "Deleting the files." >&2
    rm -rf csr.conf server-key.pem server.csr server-cert.pem
    kubectl delete secret ${certSecret} -n ${namespace}
    kubectl delete secret ${keySecret} -n ${namespace}
    kubectl delete namespace ${namespace}
    kubectl delete csr ${csrName} 2>/dev/null || true
    exit 0
fi


cat <<EOF >> csr.conf
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = SE
ST = Stockholm
L = Kista
O = Dope
OU = IT
CN = ${masterIP}

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster
DNS.5 = kubernetes.default.svc.cluster.local
DNS.6 = ${service}
DNS.7 = ${service}.${namespace}
DNS.8 = ${service}.${namespace}.svc
DNS.9 = ${service}.${namespace}.svc.cluster
DNS.10 = ${service}.${namespace}.svc.cluster.local
IP.1 = ${masterIP}

[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
EOF

openssl genrsa -out server-key.pem 2048
openssl req -new -key server-key.pem -subj "/CN=${service}.${namespace}.svc" -out server.csr -config csr.conf

# clean-up any previously created CSR for our service. Ignore errors if not present.
kubectl delete csr ${csrName} 2>/dev/null || true

# create  server cert/key CSR and  send to k8s API
cat <<EOF | kubectl create -f -
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: ${csrName}
spec:
  groups:
  - system:authenticated
  request: $(cat server.csr | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

# verify CSR has been created
while true; do
    kubectl get csr ${csrName}
    if [ "$?" -eq 0 ]; then
        break
    fi
done

# approve and fetch the signed certificate
kubectl certificate approve ${csrName}
# verify certificate has been signed
for x in $(seq 10); do
    serverCert=$(kubectl get csr ${csrName} -o jsonpath='{.status.certificate}')
    if [[ ${serverCert} != '' ]]; then
        break
    fi
    sleep 1
done
if [[ ${serverCert} == '' ]]; then
    echo "ERROR: After approving csr ${csrName}, the signed certificate did not appear on the resource. Giving up after 10 attempts." >&2
    exit 1
fi
echo ${serverCert} | openssl base64 -d -A -out server-cert.pem

# Creating namespace
kubectl create namespace ${namespace} || true

# create the secret with CA cert
kubectl create secret generic ${certSecret} \
        --from-file=cert.pem=server-cert.pem \
        --dry-run -o yaml |
    kubectl -n ${namespace} apply -f -

# create the secret with key
kubectl create secret generic ${keySecret} \
        --from-file=key.pem=server-key.pem \
        --dry-run -o yaml |
    kubectl -n ${namespace} apply -f -

# Modifying manifest.yaml, Replacing CABUNDLE and Namespace
CA_BUNDLE=$(cat server-cert.pem | base64 | tr -d '\n')

# OR you can use kubernetes extension apiserver authentication certificate
# CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')

sed -i "s/caBundle: .*$/caBundle: ${CA_BUNDLE}/g" ./manifest.yaml
sed -i "s/namespace: .*$/namespace: ${namespace}/g" ./manifest.yaml
sed -i -e "s*SERVICE_NAME*${service}*g" ./manifest.yaml
sed -i -e "s*SERVER_CERT*${certSecret}*g" ./manifest.yaml
sed -i -e "s*SERVER_KEY*${keySecret}*g" ./manifest.yaml

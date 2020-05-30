# Webhook

### This Webhook handles the addition of imagePullSecrets to ServiceAccount specified in env variable and creation of Registry Credentials Secret in namespaces which contains the credentials to pull secret from private docker registry.


### Generate certs and modifying the manifest.yaml

To run webhook on https, generate certs using script as follows.
```bash generate_certs.sh --service webhook --namespace webhook --certSecret cert --keySecret key```

To delete all the generated certs and resources do as following.
```bash generate_certs.sh --service webhook --namespace webhook --certSecret cert --keySecret key --delete true```

Deploy the webhook with manifest.
```kubectl apply -f manifest.yaml```

	Host    string `default:"0.0.0.0"`
	Port    string `default:"8888"`
	TlsCert string `default:"/etc/webhook/certs/cert.pem"`
	TlsKey  string `default:"/etc/webhook/certs/key.pem"`
	Debug   bool   `default:"false"`
	JsonLog bool   `default:"false"`
	SourceSecretName      string `required:"true" split_words:"true"`
	SourceSecretNamespace string `required:"true" split_words:"true"`
	TargetServiceAccount  string `default:"all"  split_words:"true"`

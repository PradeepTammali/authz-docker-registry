package main

import (
	"registry/authz/webhook/server"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	Host    string `default:"0.0.0.0"`
	Port    string `default:"8888"`
	TlsCert string `default:"/etc/webhook/certs/cert.pem"`
	TlsKey  string `default:"/etc/webhook/certs/key.pem"`
	Debug   bool   `default:"false"`
	JsonLog bool   `default:"false"`
}

func main() {
	// setting up logrus vars
	econfig := &Config{}
	envconfig.Process("", econfig)
	if econfig.JsonLog {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000",
			FullTimestamp:   true,
		})
	}
	if econfig.Debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug(econfig)

	// Initializing k8s config
	log.Debug("Initializing K8s Context.")
	config := k8sconfig.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Error creating k8s context...", err)
	}

	// Creating http server and starting it.
	log.Debug("Creating the server.")
	httpSrv, err := server.CreateServer(&server.ServiceAccountAdmission{K8sClient: clientset}, econfig.Host, econfig.Port, econfig.TlsCert, econfig.TlsKey)
	if err != nil {
		log.Fatal("Error while creating server...", err)
	}
	log.WithFields(log.Fields{"host": econfig.Host, "port": econfig.Port}).Info("Server is running...")
	log.Fatal(httpSrv.ListenAndServeTLS("", ""))
}

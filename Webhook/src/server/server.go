package server

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type AdmissionController struct {
	ServiceAccountAdmission *ServiceAccountAdmission
	Decoder                 runtime.Decoder
}

func (ac *AdmissionController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if data, err := ioutil.ReadAll(r.Body); err == nil {
		body = data
	}

	log.Debug("Decoding the request.")
	review := &v1beta1.AdmissionReview{}
	_, _, err := ac.Decoder.Decode(body, nil, review)
	if err != nil {
		log.Error("Can't decode request...", err)
	}

	log.Debug("Invoking the handler.")
	review = ac.ServiceAccountAdmission.HandleAdmission(review)

	log.Debug("Returining the reponse.")
	responseInBytes, err := json.Marshal(review)
	if _, err := w.Write(responseInBytes); err != nil {
		log.Error("Error while returing the response...", err)
	}
}

func CreateServer(saAdmission *ServiceAccountAdmission, host string, port string, cert string, key string) (*http.Server, error) {
	log.Debug("Loading certs.")
	cer, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		log.Error("Error while loading certs...", err)
		return nil, err
	}
	// TLS certs configurations
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		Certificates: []tls.Certificate{cer},
	}

	log.Debug("Configuring server.")
	// Adding Router, /api/health to check the health endpoint and / for handling webhook
	router := mux.NewRouter()

	// Adding health handler
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}).Methods("GET")

	// Adding Service Account handler
	acHandler := &AdmissionController{
		ServiceAccountAdmission: saAdmission,
		Decoder:                 codecs.UniversalDeserializer(),
	}
	router.Path("/").Handler(acHandler)

	httpServer := &http.Server{
		Addr:         host + ":" + port,
		Handler:      router,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	return httpServer, nil
}

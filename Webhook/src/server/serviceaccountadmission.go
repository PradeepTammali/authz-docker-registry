package server

import (
	"context"
	"time"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	SourceSecretName      string `required:"true" split_words:"true"`
	SourceSecretNamespace string `required:"true" split_words:"true"`
	TargetServiceAccount  string `default:"all" split_words:"true"`
}

type ServiceAccountAdmission struct {
	K8sClient *kubernetes.Clientset
}

func (saa *ServiceAccountAdmission) CreateRegistrySecret(namespace string, sourceSecretName string, sourceSecretNamespace string) (*corev1.Secret, error) {
	regSecret, err := saa.K8sClient.CoreV1().Secrets(sourceSecretNamespace).Get(context.TODO(), sourceSecretName, metav1.GetOptions{})
	if err != nil {
		log.Error("Error fetching the source secret...", err)
		return nil, err
	}
	// Defining secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      regSecret.Name,
			Namespace: namespace,
		},
		Data:     regSecret.Data,
		Type:     regSecret.Type,
		TypeMeta: regSecret.TypeMeta,
	}

	log.WithFields(log.Fields{"secret": sourceSecretName, "namespace": namespace}).Info("Creating secret.")
	_, err = saa.K8sClient.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		log.Error("Error creating secret...", err)
		return nil, err
	}
	return secret, nil
}

func (saa *ServiceAccountAdmission) UpdateServiceAccount(serviceAccount string, namespace string) {
	config := &Config{}
	envconfig.Process("", config)
	// Creating the Registry Credentials Secret
	log.WithFields(log.Fields{"namespace": namespace, "serviceAccount": serviceAccount}).Debug("Creating Registry Secret.")
	if config.TargetServiceAccount == "all" || config.TargetServiceAccount == serviceAccount {
		// Delay time to allow apiserver to create service account.
		log.Info("Delay 2 seconds.")
		time.Sleep(2 * time.Second)

		secret, err := saa.CreateRegistrySecret(namespace, config.SourceSecretName, config.SourceSecretNamespace)
		if err != nil {
			log.Error("Error creating Registry Credentials Secret...", err)
			return
		}

		log.WithFields(log.Fields{"namespace": namespace, "serviceAccount": serviceAccount}).Debug("Fetching service account.")
		// Updating the Service Account.
		sa, err := saa.K8sClient.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), serviceAccount, metav1.GetOptions{})
		if err != nil {
			log.Error("Error fetching service account...", err)
			return
		}
		sa.ImagePullSecrets = []corev1.LocalObjectReference{{Name: secret.Name}}
		_, err = saa.K8sClient.CoreV1().ServiceAccounts(namespace).Update(context.TODO(), sa, metav1.UpdateOptions{})
		if err != nil {
			log.Error("Error updating service account.", err)
			return
		}
		log.WithFields(log.Fields{"serviceaccount": serviceAccount, "namespace": namespace}).Info("Service Account updated.")
	} else {
		log.WithFields(log.Fields{"serviceAccount": serviceAccount, "targetServiceAccount": config.TargetServiceAccount}).Info("Skipping...")
	}
	return
}

func (saa *ServiceAccountAdmission) HandleAdmission(review *v1beta1.AdmissionReview) *v1beta1.AdmissionReview {
	log.Debug("Invoking UpdateServiceAccount with go routines.")
	go saa.UpdateServiceAccount(review.Request.Name, review.Request.Namespace)
	// returing the response
	review.Response = &v1beta1.AdmissionResponse{
		Allowed: true,
		Result: &metav1.Status{
			Message: "Request processed and validated.",
			Code:    200,
			Status:  "Success",
		},
	}
	return review
}

package admission

import (
	"crypto/tls"

	"k8s.io/client-go/kubernetes"

	configurationv1alpha1 "github.com/aledbf/ingress-experiments/internal/apis/clientset/versioned"
)

// NewValidatingAdmissionWebhook creates a ValidatingAdmissionWebhook struct
// that will use the specified client to access the API.
func NewValidatingAdmissionWebhook(
	namespace string,
	kubeClient kubernetes.Interface,
	confClient configurationv1alpha1.Interface) *ValidatingAdmissionWebhook {
	return &ValidatingAdmissionWebhook{
		namespace:  namespace,
		kubeClient: kubeClient,
		confClient: confClient,
	}
}

// ValidatingAdmissionWebhook represents a validating admission webhook.
type ValidatingAdmissionWebhook struct {
	namespace      string
	kubeClient     kubernetes.Interface
	confClient     configurationv1alpha1.Interface
	tlsCertificate tls.Certificate
}

// Register registers the validating admission webhook.
func (s *ValidatingAdmissionWebhook) Register() error {
	return nil
}

func (s *ValidatingAdmissionWebhook) Run(stopCh chan struct{}) {
}

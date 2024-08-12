package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateServiceForYellow(clientset *kubernetes.Clientset) utils.Result {
	service, err := clientset.CoreV1().Services("colors").Get(context.TODO(), "yellow-service", metav1.GetOptions{})

	passed := err == nil &&
		service.Spec.Ports[0].Port == 80 &&
		service.Spec.Ports[0].TargetPort.IntVal == 3000 &&
		service.Spec.Selector["app"] == "yellow-deployment"

	return utils.Result{
		TestName:   "Question 21 - Create a service yellow-service for the deployment yellow-deployment in namespace colors with port 80 and target port 3000",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

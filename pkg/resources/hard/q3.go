package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeploymentAndService(clientset *kubernetes.Clientset) utils.Result {
	deployment, err := clientset.AppsV1().Deployments("latam").Get(context.TODO(), "redis", metav1.GetOptions{})
	service, err := clientset.CoreV1().Services("latam").Get(context.TODO(), "redis-service", metav1.GetOptions{})

	passed := err == nil &&
		service != nil &&
		deployment.Name == "redis" &&
		service.Name == "redis-service" &&
		service.Spec.Ports[0].Port == 6379 &&
		deployment.Spec.Template.Spec.Containers[0].Image == "redis:alpine"

	return utils.Result{
		TestName:   "Question 3 - Create a deployment redis with redis:alpine image and a service named redis-service on port 6379 in namespace latam",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

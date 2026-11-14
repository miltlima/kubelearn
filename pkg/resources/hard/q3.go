package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeploymentAndService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deployment, deployErr := clientset.AppsV1().Deployments("latam").Get(ctx, "redis", metav1.GetOptions{})
	service, serviceErr := clientset.CoreV1().Services("latam").Get(ctx, "redis-service", metav1.GetOptions{})

	var report utils.Report
	if report.Found(deployErr, "Deployment", "redis", "latam") {
		report.ExpectImage(deployment.Spec.Template.Spec, "redis:alpine")
	}
	if report.Found(serviceErr, "Service", "redis-service", "latam") {
		report.Expect(len(service.Spec.Ports) > 0 && service.Spec.Ports[0].Port == 6379,
			"The service's first port should be 6379.")
	}
	return report.Result("Question 3 - Create a deployment redis with redis:alpine image and a service named redis-service on port 6379 in namespace latam", "Hard")
}

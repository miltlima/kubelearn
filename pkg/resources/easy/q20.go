package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeploymentYellow(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deployment, err := clientset.AppsV1().Deployments("colors").Get(ctx, "yellow-deployment", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "yellow-deployment", "colors") {
		report.ExpectImage(deployment.Spec.Template.Spec, "bonovoo/node-app:1.0")
		replicas := deployment.Spec.Replicas
		report.Expect(utils.PtrEquals(replicas, 2), "Replicas is %s, expected 2.", utils.ShownPtr(replicas))
	}
	return report.Result("Question 20 - Create a deployment yellow-deployment with bonovoo/node-app:1.0 image and 2 replicas in namespace colors", "Easy")
}

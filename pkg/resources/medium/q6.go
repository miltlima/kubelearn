package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateLabel(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods("asia").Get(ctx, "tshoot", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "tshoot", "asia") {
		report.ExpectImage(pod.Spec, "amazon/amazon-ecs-network-sidecar:latest")
		country := pod.ObjectMeta.Labels["country"]
		report.Expect(country == "china", "Label country is %s, expected china.", utils.Shown(country))
	}
	return report.Result("Question 6 - Create a pod tshoot with label country=china with amazon/amazon-ecs-network-sidecar:latest image in namespace asia", "Medium")
}

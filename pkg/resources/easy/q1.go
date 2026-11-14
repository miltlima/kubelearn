package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods("default").Get(ctx, "nginx", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "nginx", "default") {
		report.ExpectImage(pod.Spec, "nginx:alpine")
	}
	return report.Result("Question 1 - Create a pod nginx name with nginx:alpine image", "Easy")
}

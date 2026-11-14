package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckPodError(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods("bandai").Get(ctx, "gundamv", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "gundamv", "bandai") {
		// Deliberately vague: finding the cause is the exercise.
		container := utils.FirstContainer(pod.Spec)
		report.Expect(container != nil && container.Image == "nginx:alpine",
			"The pod is not fixed yet. Find out why its container cannot start.")
	}
	return report.Result("Question 10 - Identify and fix the issue in the pod gundamv in namespace bandai", "Hard")
}

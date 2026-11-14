package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePodAddSecret(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, podErr := clientset.CoreV1().Pods("colors").Get(ctx, "purple", metav1.GetOptions{})
	secret, secretErr := clientset.CoreV1().Secrets("colors").Get(ctx, "secret-purple", metav1.GetOptions{})

	var report utils.Report
	if report.Found(secretErr, "Secret", "secret-purple", "colors") {
		singer := string(secret.Data["singer"])
		report.Expect(singer == "prince", "Secret key singer is %s, expected prince.", utils.Shown(singer))
	}
	if report.Found(podErr, "Pod", "purple", "colors") {
		report.ExpectImage(pod.Spec, "redis:alpine")
		report.Expect(len(pod.Spec.Volumes) > 0 && pod.Spec.Volumes[0].Secret != nil &&
			pod.Spec.Volumes[0].Secret.SecretName == "secret-purple",
			"The pod's first volume should mount secret secret-purple.")
	}
	return report.Result("Question 13 - Add a secret secret-purple with data singer=prince to the pod purple with image redis:alpine in colors namespace", "Hard")
}

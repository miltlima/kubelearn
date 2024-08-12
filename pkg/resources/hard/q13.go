package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePodAddSecret(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("colors").Get(context.TODO(), "purple", metav1.GetOptions{})
	secret, err := clientset.CoreV1().Secrets("colors").Get(context.TODO(), "secret-purple", metav1.GetOptions{})

	passed := err == nil &&
		pod.Spec.Volumes[0].Secret.SecretName == "secret-purple" &&
		string(secret.Data["singer"]) == "prince" &&
		pod.Spec.Containers[0].Image == "redis:alpine"

	return utils.Result{
		TestName:   "Question 13 - Add a secret secret-purple with data singer=prince to the pod purple with image redis:alpine in colors namespace",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

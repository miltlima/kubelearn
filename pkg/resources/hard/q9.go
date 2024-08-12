package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePodVolumeClaim(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("public").Get(context.TODO(), "webserver", metav1.GetOptions{})

	passed := err == nil &&
		pod.Spec.Containers[0].Image == "nginx:alpine" &&
		pod.Spec.Volumes[0].PersistentVolumeClaim.ClaimName == "unicorn-pvc" &&
		pod.Spec.Containers[0].VolumeMounts[0].MountPath == "/usr/share/nginx/html" &&
		pod.Spec.Volumes[0].Name == "unicorn-pv"

	return utils.Result{
		TestName:   "Question 9 - Create a pod webserver in public namespace with nginx:alpine image, volume mount /usr/share/nginx/html, and a persistent volume claim unicorn-pvc",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

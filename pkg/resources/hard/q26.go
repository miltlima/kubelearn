package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateStatefulSet(clientset *kubernetes.Clientset) utils.Result {
	statefulset, err := clientset.AppsV1().StatefulSets("default").Get(context.TODO(), "statefulset-gain", metav1.GetOptions{})

	passed := err == nil &&
		statefulset.Name == "statefulset-gain" &&
		statefulset.Spec.Template.Spec.Containers[0].Image == "busybox:1.28" &&
		statefulset.Spec.Template.Spec.Containers[0].Command[0] == "sleep 3600" &&
		statefulset.Status.ReadyReplicas == 3

	return utils.Result{
		TestName:   "Question 26 - Create a statefulset statefulset-gain with image busybox:1.28, command 'sleep 3600', and 3 replicas",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

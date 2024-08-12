package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePersistentVolume(clientset *kubernetes.Clientset) utils.Result {
	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), "unicorn-pv", metav1.GetOptions{})
	passed := err == nil && pv.Spec.Capacity.Storage().String() == "1Gi" && pv.Spec.AccessModes[0] == "ReadWriteMany" && pv.Spec.HostPath.Path == "/tmp/data"

	return utils.Result{
		TestName:   "Question 7 - Create a persistent volume unicorn-pv with capacity 1Gi, access mode ReadWriteMany, and host path /tmp/data",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

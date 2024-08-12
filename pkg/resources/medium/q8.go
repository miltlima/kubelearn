package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePersistentVolumeClaim(clientset *kubernetes.Clientset) utils.Result {
	pvc, err := clientset.CoreV1().PersistentVolumeClaims("default").Get(context.TODO(), "unicorn-pvc", metav1.GetOptions{})
	passed := err == nil && pvc.Spec.Resources.Requests.Storage().String() == "400Mi" && pvc.Spec.AccessModes[0] == "ReadWriteMany"

	return utils.Result{
		TestName:   "Question 8 - Create a persistent volume claim unicorn-pvc with capacity 400Mi and access mode ReadWriteMany",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

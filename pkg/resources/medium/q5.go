package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateConfigMap(clientset *kubernetes.Clientset) utils.Result {
	configMap, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "europe-configmap", metav1.GetOptions{})
	passed := err == nil && configMap.Name == "europe-configmap" && configMap.Data["France"] == "Paris"

	return utils.Result{
		TestName:   "Question 5 - Create a configmap europe-configmap with data France=Paris",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

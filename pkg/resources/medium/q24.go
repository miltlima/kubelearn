package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateJob(clientset *kubernetes.Clientset) utils.Result {
	job, err := clientset.BatchV1().Jobs("default").Get(context.TODO(), "job-gain", metav1.GetOptions{})
	passed := err == nil &&
		*job.Spec.Parallelism == 2 &&
		*job.Spec.Completions == 4 &&
		*job.Spec.BackoffLimit == 3 &&
		*job.Spec.ActiveDeadlineSeconds == 40

	return utils.Result{
		TestName:   "Question 24 - Create a job job-gain with parallelism 2, completions 4, backoffLimit 3, and deadlineSeconds 40",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

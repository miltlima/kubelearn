package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateCronjob(clientset *kubernetes.Clientset) utils.Result {
	cronjob, err := clientset.BatchV1().CronJobs("default").Get(context.TODO(), "cronjob-gain", metav1.GetOptions{})
	passed := err == nil &&
		cronjob.Spec.Schedule == "*/5 * * * *" &&
		cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image == "busybox:1.28" &&
		cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0] == "sleep 3600" &&
		cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy == "Never"

	return utils.Result{
		TestName:   "Question 25 - Create a cronjob cronjob-gain to run every 5 minutes with image busybox:1.28, command 'sleep 3600', and restartPolicy Never",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

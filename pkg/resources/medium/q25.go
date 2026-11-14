package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateCronjob(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	cronjob, err := clientset.BatchV1().CronJobs("default").Get(ctx, "cronjob-gain", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "CronJob", "cronjob-gain", "default") {
		podSpec := cronjob.Spec.JobTemplate.Spec.Template.Spec
		report.Expect(cronjob.Spec.Schedule == "*/5 * * * *", "Schedule is %s, expected */5 * * * *.", utils.Shown(cronjob.Spec.Schedule))
		container := report.ExpectImage(podSpec, "busybox:1.28")
		report.Expect(utils.RunsCommand(container, "sleep 3600"), "The container should run the command sleep 3600.")
		report.Expect(podSpec.RestartPolicy == "Never", "Restart policy is %s, expected Never.", utils.Shown(string(podSpec.RestartPolicy)))
	}
	return report.Result("Question 25 - Create a cronjob cronjob-gain to run every 5 minutes with image busybox:1.28, command 'sleep 3600', and restartPolicy Never", "Medium")
}

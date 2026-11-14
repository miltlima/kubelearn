package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateJob(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	job, err := clientset.BatchV1().Jobs("default").Get(ctx, "job-gain", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Job", "job-gain", "default") {
		spec := job.Spec
		report.Expect(utils.PtrEquals(spec.Parallelism, 2), "Parallelism is %s, expected 2.", utils.ShownPtr(spec.Parallelism))
		report.Expect(utils.PtrEquals(spec.Completions, 4), "Completions is %s, expected 4.", utils.ShownPtr(spec.Completions))
		report.Expect(utils.PtrEquals(spec.BackoffLimit, 3), "Backoff limit is %s, expected 3.", utils.ShownPtr(spec.BackoffLimit))
		report.Expect(utils.PtrEquals(spec.ActiveDeadlineSeconds, 40), "Active deadline is %s, expected 40 seconds.", utils.ShownPtr(spec.ActiveDeadlineSeconds))
	}
	return report.Result("Question 24 - Create a job job-gain with parallelism 2, completions 4, backoffLimit 3, and deadlineSeconds 40", "Medium")
}

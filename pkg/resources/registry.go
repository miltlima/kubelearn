// Package resources is the single registry of KubeLearn questions.
package resources

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"kubelearn/pkg/resources/easy"
	"kubelearn/pkg/resources/hard"
	"kubelearn/pkg/resources/medium"
	"kubelearn/pkg/utils"
)

// Check inspects the cluster and reports whether one question is solved.
type Check func(context.Context, kubernetes.Interface) utils.Result

// Checks lists every question in order: Checks[0] is question 1.
// To add a question, write its checker and append it here.
var Checks = []Check{
	easy.CreatePod,                       // Question 1
	medium.CreateDeployment,              // Question 2
	hard.CreateDeploymentAndService,      // Question 3
	easy.CreateNamespace,                 // Question 4
	medium.CreateConfigMap,               // Question 5
	medium.CreateLabel,                   // Question 6
	medium.CreatePersistentVolume,        // Question 7
	medium.CreatePersistentVolumeClaim,   // Question 8
	hard.CreatePodVolumeClaim,            // Question 9
	hard.CheckPodError,                   // Question 10
	hard.CreateNetPolRule,                // Question 11
	easy.CreateSecret,                    // Question 12
	hard.CreatePodAddSecret,              // Question 13
	easy.CreateServiceAccount,            // Question 14
	medium.AddServiceAccountToDeployment, // Question 15
	medium.ChangeReplicaCount,            // Question 16
	medium.CreateHpa,                     // Question 17
	medium.AddSecurityContext,            // Question 18
	medium.AddLivenessProbe,              // Question 19
	easy.CreateDeploymentYellow,          // Question 20
	hard.CreateServiceForYellow,          // Question 21
	hard.CreateIngressYellow,             // Question 22
	hard.CreateRoleOne,                   // Question 23
	medium.CreateJob,                     // Question 24
	medium.CreateCronjob,                 // Question 25
	hard.CreateStatefulSet,               // Question 26
}

// Evaluate runs every check against the cluster, in question order.
//
// It stops and returns the context error if ctx is cancelled or its deadline
// passes, so an unreachable cluster is reported as an error instead of as
// every question failing.
func Evaluate(ctx context.Context, clientset kubernetes.Interface) ([]utils.Result, error) {
	results := make([]utils.Result, 0, len(Checks))
	for _, check := range Checks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		results = append(results, check(ctx, clientset))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Score returns the percentage (0-100) of passed results.
func Score(results []utils.Result) float64 {
	if len(results) == 0 {
		return 0
	}
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	return float64(passed) / float64(len(results)) * 100
}

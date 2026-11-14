// Package resources is the single registry of KubeLearn questions.
package resources

import (
	"context"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"kubelearn/pkg/resources/ckne"
	"kubelearn/pkg/resources/cks"
	"kubelearn/pkg/resources/easy"
	"kubelearn/pkg/resources/hard"
	"kubelearn/pkg/resources/medium"
	"kubelearn/pkg/utils"
)

// Tracks group the questions by the exam they prepare for.
const (
	TrackCore = "Core" // general Kubernetes practice
	TrackCKS  = "CKS"  // Certified Kubernetes Security Specialist
	TrackCKNE = "CKNE" // Certified Kubernetes Network Engineer
)

// Tracks lists the tracks in display order.
var Tracks = []string{TrackCore, TrackCKS, TrackCKNE}

// Clients gives the checks access to the cluster.
type Clients struct {
	Kube    kubernetes.Interface // built-in resources
	Dynamic dynamic.Interface    // custom resources, such as the Gateway API
}

// Check inspects the cluster and reports whether one question is solved.
type Check func(context.Context, Clients) utils.Result

// Question is one registered question.
type Question struct {
	Track string
	Check Check
}

// typed registers a check that reads built-in resources.
func typed(track string, check func(context.Context, kubernetes.Interface) utils.Result) Question {
	return Question{track, func(ctx context.Context, clients Clients) utils.Result { return check(ctx, clients.Kube) }}
}

// custom registers a check that reads custom resources.
func custom(track string, check func(context.Context, dynamic.Interface) utils.Result) Question {
	return Question{track, func(ctx context.Context, clients Clients) utils.Result { return check(ctx, clients.Dynamic) }}
}

// Questions lists every question in order: Questions[0] is question 1.
// To add a question, write its checker and append it here.
var Questions = []Question{
	typed(TrackCore, easy.CreatePod),                       // Question 1
	typed(TrackCore, medium.CreateDeployment),              // Question 2
	typed(TrackCore, hard.CreateDeploymentAndService),      // Question 3
	typed(TrackCore, easy.CreateNamespace),                 // Question 4
	typed(TrackCore, medium.CreateConfigMap),               // Question 5
	typed(TrackCore, medium.CreateLabel),                   // Question 6
	typed(TrackCore, medium.CreatePersistentVolume),        // Question 7
	typed(TrackCore, medium.CreatePersistentVolumeClaim),   // Question 8
	typed(TrackCore, hard.CreatePodVolumeClaim),            // Question 9
	typed(TrackCore, hard.CheckPodError),                   // Question 10
	typed(TrackCore, hard.CreateNetPolRule),                // Question 11
	typed(TrackCore, easy.CreateSecret),                    // Question 12
	typed(TrackCore, hard.CreatePodAddSecret),              // Question 13
	typed(TrackCore, easy.CreateServiceAccount),            // Question 14
	typed(TrackCore, medium.AddServiceAccountToDeployment), // Question 15
	typed(TrackCore, medium.ChangeReplicaCount),            // Question 16
	typed(TrackCore, medium.CreateHpa),                     // Question 17
	typed(TrackCore, medium.AddSecurityContext),            // Question 18
	typed(TrackCore, medium.AddLivenessProbe),              // Question 19
	typed(TrackCore, easy.CreateDeploymentYellow),          // Question 20
	typed(TrackCore, hard.CreateServiceForYellow),          // Question 21
	typed(TrackCore, hard.CreateIngressYellow),             // Question 22
	typed(TrackCore, hard.CreateRoleOne),                   // Question 23
	typed(TrackCore, medium.CreateJob),                     // Question 24
	typed(TrackCore, medium.CreateCronjob),                 // Question 25
	typed(TrackCore, hard.CreateStatefulSet),               // Question 26

	typed(TrackCKS, cks.PodSecurityNamespaces),      // Question 27
	typed(TrackCKS, cks.SeccompProfile),             // Question 28
	typed(TrackCKS, cks.AppArmorProfile),            // Question 29
	typed(TrackCKS, cks.StrictSecurityContext),      // Question 30
	typed(TrackCKS, cks.RestrictEgress),             // Question 31
	typed(TrackCKS, cks.ServiceAccountWithoutToken), // Question 32
	typed(TrackCKS, cks.LeastPrivilegeRole),         // Question 33
	typed(TrackCKS, cks.RemoveClusterAdmin),         // Question 34
	typed(TrackCKS, cks.RemovePrivileges),           // Question 35
	typed(TrackCKS, cks.ImmutableSecret),            // Question 36
	typed(TrackCKS, cks.SandboxedRuntime),           // Question 37

	typed(TrackCKNE, ckne.NodePortService),           // Question 38
	typed(TrackCKNE, ckne.HeadlessService),           // Question 39
	typed(TrackCKNE, ckne.ExternalNameService),       // Question 40
	typed(TrackCKNE, ckne.SessionAffinityService),    // Question 41
	typed(TrackCKNE, ckne.LocalTrafficPolicyService), // Question 42
	typed(TrackCKNE, ckne.ReadinessProbe),            // Question 43
	typed(TrackCKNE, ckne.CoreDNSRewrite),            // Question 44
	typed(TrackCKNE, ckne.PodDNSConfig),              // Question 45
	typed(TrackCKNE, ckne.DefaultDenyIngress),        // Question 46
	typed(TrackCKNE, ckne.IngressFromNamespace),      // Question 47
	typed(TrackCKNE, ckne.EgressExceptPrivate),       // Question 48
	typed(TrackCKNE, ckne.TLSSecret),                 // Question 49
	typed(TrackCKNE, ckne.IngressWithTLS),            // Question 50
	custom(TrackCKNE, ckne.HTTPGateway),              // Question 51
	custom(TrackCKNE, ckne.PathRouting),              // Question 52
	custom(TrackCKNE, ckne.TrafficSplit),             // Question 53
	custom(TrackCKNE, ckne.HTTPSListener),            // Question 54
}

// ParseTrack resolves a track name, ignoring case. An empty name or "all"
// means every track and returns "". ok is false for an unknown name.
func ParseTrack(name string) (track string, ok bool) {
	if name == "" || strings.EqualFold(name, "all") {
		return "", true
	}
	for _, known := range Tracks {
		if strings.EqualFold(name, known) {
			return known, true
		}
	}
	return "", false
}

// Difficulties lists the difficulty levels in ascending order.
var Difficulties = []string{"Easy", "Medium", "Hard"}

// ParseDifficulty resolves a difficulty name, ignoring case. An empty name or
// "all" means every level and returns "". ok is false for an unknown name.
func ParseDifficulty(name string) (difficulty string, ok bool) {
	if name == "" || strings.EqualFold(name, "all") {
		return "", true
	}
	for _, known := range Difficulties {
		if strings.EqualFold(name, known) {
			return known, true
		}
	}
	return "", false
}

// Filter selects which questions to evaluate. Empty fields select everything.
type Filter struct {
	Track      string
	Difficulty string
}

// Evaluate runs the checks selected by filter against the cluster, in
// question order.
//
// It stops and returns the context error if ctx is cancelled or its deadline
// passes, so an unreachable cluster is reported as an error instead of as
// every question failing.
func Evaluate(ctx context.Context, clients Clients, filter Filter) ([]utils.Result, error) {
	// The difficulty is declared inside each check, so it is looked up in the
	// catalog to skip a question without running it.
	var described []utils.Result
	if filter.Difficulty != "" {
		described = Catalog()
	}

	results := make([]utils.Result, 0, len(Questions))
	for i, question := range Questions {
		if filter.Track != "" && question.Track != filter.Track {
			continue
		}
		if filter.Difficulty != "" && described[i].Difficulty != filter.Difficulty {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result := question.Check(ctx, clients)
		result.Track = question.Track
		results = append(results, result)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

var (
	catalogOnce sync.Once
	catalog     []utils.Result
)

// Catalog lists every question (name, difficulty and track) without touching
// a cluster. The names live inside the checks, so they are read by running
// each check once against an empty in-memory cluster.
func Catalog() []utils.Result {
	catalogOnce.Do(func() {
		empty := Clients{
			Kube:    fake.NewSimpleClientset(),
			Dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
		}
		results, _ := Evaluate(context.Background(), empty, Filter{})
		for i := range results {
			results[i].Passed = false
			results[i].Feedback = ""
		}
		catalog = results
	})
	return append([]utils.Result(nil), catalog...)
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

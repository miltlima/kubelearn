package utils

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Report collects the reasons a question is not solved yet. A question passes
// when its report has no problems.
type Report struct {
	problems []string
}

// Expect records a problem when ok is false, and returns ok.
func (r *Report) Expect(ok bool, format string, args ...any) bool {
	if !ok {
		r.problems = append(r.problems, fmt.Sprintf(format, args...))
	}
	return ok
}

// Found checks the error from fetching an object and returns whether the
// object can be inspected. A missing object and a failed request are reported
// differently, so a cluster problem is not mistaken for a wrong answer.
// Pass an empty namespace for cluster-scoped objects.
func (r *Report) Found(err error, kind, name, namespace string) bool {
	switch {
	case err == nil:
		return true
	case apierrors.IsNotFound(err) && namespace == "":
		r.Expect(false, "%s %s not found.", kind, name)
	case apierrors.IsNotFound(err):
		r.Expect(false, "%s %s not found in namespace %s.", kind, name, namespace)
	default:
		r.Expect(false, "Could not read %s %s from the cluster: %v.", strings.ToLower(kind), name, err)
	}
	return false
}

// ExpectImage checks the image of the first container of a pod spec. It
// returns that container, or nil when the spec has none.
func (r *Report) ExpectImage(spec corev1.PodSpec, want string) *corev1.Container {
	container := FirstContainer(spec)
	if !r.Expect(container != nil, "No container defined.") {
		return nil
	}
	r.Expect(container.Image == want, "Container image is %s, expected %s.", Shown(container.Image), want)
	return container
}

// Result builds the outcome of a question from the collected problems.
func (r *Report) Result(testName, difficulty string) Result {
	return Result{
		TestName:   testName,
		Passed:     len(r.problems) == 0,
		Difficulty: difficulty,
		Feedback:   strings.Join(r.problems, " "),
	}
}

// Shown formats a found value for feedback, making an empty one visible.
func Shown(value string) string {
	if value == "" {
		return "not set"
	}
	return value
}

// ShownPtr formats an optional found value for feedback.
func ShownPtr[T any](p *T) string {
	if p == nil {
		return "not set"
	}
	return fmt.Sprint(*p)
}

// RunsCommand reports whether the container's command line (command plus
// args) contains want, however it was split: ["sleep", "3600"],
// ["sh", "-c", "sleep 3600"] and ["sleep 3600"] all run "sleep 3600".
func RunsCommand(container *corev1.Container, want string) bool {
	if container == nil {
		return false
	}
	parts := append(append([]string{}, container.Command...), container.Args...)
	return strings.Contains(strings.Join(parts, " "), want)
}

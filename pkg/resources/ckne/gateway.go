package ckne

import (
	"context"

	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Gateway API objects are custom resources, so they are read with the dynamic
// client as generic maps instead of typed structs.
var (
	gatewayResource   = schema.GroupVersionResource{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways"}
	httpRouteResource = schema.GroupVersionResource{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes"}
)

// items returns the list of objects at the given path, skipping anything that
// is not an object.
func items(object map[string]any, path ...string) []map[string]any {
	list, _, _ := unstructured.NestedSlice(object, path...)
	var result []map[string]any
	for _, item := range list {
		if entry, ok := item.(map[string]any); ok {
			result = append(result, entry)
		}
	}
	return result
}

func text(object map[string]any, path ...string) string {
	value, _, _ := unstructured.NestedString(object, path...)
	return value
}

// number reads an integer field; ok is false when it is missing.
func number(object map[string]any, path ...string) (value int64, ok bool) {
	raw, found, _ := unstructured.NestedFieldNoCopy(object, path...)
	if !found {
		return 0, false
	}
	switch n := raw.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	case float64:
		return int64(n), true
	}
	return 0, false
}

func isNumber(object map[string]any, want int64, path ...string) bool {
	value, ok := number(object, path...)
	return ok && value == want
}

func texts(object map[string]any, path ...string) []string {
	values, _, _ := unstructured.NestedStringSlice(object, path...)
	return values
}

func attachedTo(route map[string]any, gateway string) bool {
	for _, parent := range items(route, "spec", "parentRefs") {
		if text(parent, "name") == gateway {
			return true
		}
	}
	return false
}

// backend finds the backend reference to a service in a route rule.
func backend(rule map[string]any, service string) map[string]any {
	for _, ref := range items(rule, "backendRefs") {
		if text(ref, "name") == service {
			return ref
		}
	}
	return nil
}

func sendsTo(rule map[string]any, service string, port int64) bool {
	ref := backend(rule, service)
	return ref != nil && isNumber(ref, port, "port")
}

// matchesPrefix reports whether the rule matches the path prefix. A rule
// without matches is a catch-all, which is the same as prefix "/".
func matchesPrefix(rule map[string]any, prefix string) bool {
	matches := items(rule, "matches")
	if len(matches) == 0 {
		return prefix == "/"
	}
	for _, match := range matches {
		kind := text(match, "path", "type")
		if (kind == "" || kind == "PathPrefix") && text(match, "path", "value") == prefix {
			return true
		}
	}
	return false
}

func HTTPGateway(ctx context.Context, client dynamic.Interface) utils.Result {
	gateway, err := client.Resource(gatewayResource).Namespace(netlab).Get(ctx, "web-gateway", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Gateway", "web-gateway", netlab) {
		class := text(gateway.Object, "spec", "gatewayClassName")
		report.Expect(class == "kubelearn", "Gateway class is %s, expected kubelearn.", utils.Shown(class))

		listening := false
		for _, listener := range items(gateway.Object, "spec", "listeners") {
			listening = listening || (text(listener, "protocol") == "HTTP" && isNumber(listener, 80, "port"))
		}
		report.Expect(listening, "No listener accepts HTTP on port 80.")
	}
	return report.Result("Question 51 - Create gateway web-gateway in namespace netlab with gateway class kubelearn and an HTTP listener on port 80", "Medium")
}

func PathRouting(ctx context.Context, client dynamic.Interface) utils.Result {
	route, err := client.Resource(httpRouteResource).Namespace(netlab).Get(ctx, "web-route", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "HTTPRoute", "web-route", netlab) {
		report.Expect(attachedTo(route.Object, "web-gateway"), "The route is not attached to gateway web-gateway.")
		report.Expect(utils.Contains(texts(route.Object, "spec", "hostnames"), "shop.example.com"), "Hostname shop.example.com is missing.")

		api, fallback := false, false
		for _, rule := range items(route.Object, "spec", "rules") {
			api = api || (matchesPrefix(rule, "/api") && sendsTo(rule, "api", 8080))
			fallback = fallback || (matchesPrefix(rule, "/") && sendsTo(rule, "web", 80))
		}
		report.Expect(api, "No rule sends path prefix /api to service api on port 8080.")
		report.Expect(fallback, "No rule sends the remaining traffic to service web on port 80.")
	}
	return report.Result("Question 52 - Create HTTP route web-route in namespace netlab on gateway web-gateway for host shop.example.com, sending path prefix /api to service api on port 8080 and everything else to service web on port 80", "Hard")
}

func TrafficSplit(ctx context.Context, client dynamic.Interface) utils.Result {
	route, err := client.Resource(httpRouteResource).Namespace(netlab).Get(ctx, "web-canary", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "HTTPRoute", "web-canary", netlab) {
		report.Expect(attachedTo(route.Object, "web-gateway"), "The route is not attached to gateway web-gateway.")
		report.Expect(utils.Contains(texts(route.Object, "spec", "hostnames"), "canary.example.com"), "Hostname canary.example.com is missing.")

		split := false
		for _, rule := range items(route.Object, "spec", "rules") {
			stable, canary := backend(rule, "web"), backend(rule, "web-v2")
			split = split || (stable != nil && canary != nil &&
				isNumber(stable, 80, "port") && isNumber(canary, 80, "port") &&
				isNumber(stable, 90, "weight") && isNumber(canary, 10, "weight"))
		}
		report.Expect(split, "No rule splits traffic 90 to service web and 10 to service web-v2 on port 80.")
	}
	return report.Result("Question 53 - Create HTTP route web-canary in namespace netlab on gateway web-gateway for host canary.example.com that sends 90% of the traffic to service web and 10% to service web-v2, both on port 80", "Hard")
}

func HTTPSListener(ctx context.Context, client dynamic.Interface) utils.Result {
	gateway, err := client.Resource(gatewayResource).Namespace(netlab).Get(ctx, "web-gateway", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Gateway", "web-gateway", netlab) {
		var https map[string]any
		for _, listener := range items(gateway.Object, "spec", "listeners") {
			if text(listener, "protocol") == "HTTPS" && isNumber(listener, 443, "port") {
				https = listener
			}
		}
		if report.Expect(https != nil, "No listener accepts HTTPS on port 443.") {
			hostname := text(https, "hostname")
			report.Expect(hostname == "shop.example.com", "The HTTPS listener hostname is %s, expected shop.example.com.", utils.Shown(hostname))
			mode := text(https, "tls", "mode")
			report.Expect(mode == "" || mode == "Terminate", "TLS mode is %s, expected Terminate.", mode)

			certificate := false
			for _, ref := range items(https, "tls", "certificateRefs") {
				certificate = certificate || text(ref, "name") == "shop-tls"
			}
			report.Expect(certificate, "The HTTPS listener does not use secret shop-tls as its certificate.")
		}
	}
	return report.Result("Question 54 - Add an HTTPS listener on port 443 for host shop.example.com to gateway web-gateway in namespace netlab, terminating TLS with secret shop-tls", "Hard")
}

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"kubelearn"
	"kubelearn/pkg/resources"
	"kubelearn/pkg/utils"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("/terminal", handleWebSocketTerminal)
	return httptest.NewServer(withCORS(mux))
}

func TestParseAllowedOrigins(t *testing.T) {
	defaults := parseAllowedOrigins("")
	if !defaults["http://localhost:3000"] || !defaults["http://127.0.0.1:3000"] || len(defaults) != 4 {
		t.Errorf("unexpected defaults: %v", defaults)
	}

	custom := parseAllowedOrigins(" http://localhost:3001/ ,, http://example.test ")
	if !custom["http://localhost:3001"] || !custom["http://example.test"] || custom["http://localhost:3000"] || len(custom) != 4 {
		t.Errorf("unexpected custom origins: %v", custom)
	}
}

func TestCORSByOrigin(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	cases := []struct {
		origin     string
		wantStatus int
		wantHeader string
	}{
		{"http://localhost:3000", http.StatusOK, "http://localhost:3000"},
		{"", http.StatusOK, ""},
		{"http://evil.example", http.StatusForbidden, ""},
		{"http://localhost:3000.evil.example", http.StatusForbidden, ""},
	}
	for _, tc := range cases {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/ping", nil)
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("origin %q: %v", tc.origin, err)
		}
		resp.Body.Close()
		if resp.StatusCode != tc.wantStatus {
			t.Errorf("origin %q: status %d, want %d", tc.origin, resp.StatusCode, tc.wantStatus)
		}
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != tc.wantHeader {
			t.Errorf("origin %q: Access-Control-Allow-Origin %q, want %q", tc.origin, got, tc.wantHeader)
		}
	}
}

func TestTerminalRejectsForeignOrigins(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/terminal"

	for _, origin := range []string{"", "http://evil.example", "null"} {
		header := http.Header{}
		if origin != "" {
			header.Set("Origin", origin)
		}
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
		if err == nil {
			conn.Close()
			t.Errorf("origin %q: terminal connection was accepted", origin)
			continue
		}
		if resp == nil || resp.StatusCode != http.StatusForbidden {
			t.Errorf("origin %q: expected 403, got %v (%v)", origin, resp, err)
		}
	}
}

func TestTerminalAcceptsFrontendOrigin(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/terminal"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {"http://localhost:3000"}})
	if err != nil {
		t.Fatalf("frontend origin rejected: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("echo kubelearn-ok\n")); err != nil {
		t.Fatal(err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(msg)) != "kubelearn-ok" {
		t.Errorf("unexpected terminal output: %q", msg)
	}
}

// stubProvision replaces provisionLab for the test. Each run is announced on
// started and blocks until release is closed.
func stubProvision(t *testing.T) (started chan struct{}, release chan struct{}) {
	t.Helper()
	started = make(chan struct{}, 10)
	release = make(chan struct{})
	original := provisionLab
	provisionLab = func(context.Context) error {
		started <- struct{}{}
		<-release
		return nil
	}
	t.Cleanup(func() { provisionLab = original })
	return started, release
}

func requestSetup(method string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	setupEnvironment(rec, httptest.NewRequest(method, "/setup", nil))
	return rec
}

func TestSetupRejectsGet(t *testing.T) {
	started, _ := stubProvision(t)
	if rec := requestSetup(http.MethodGet); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /setup: status %d, want 405", rec.Code)
	}
	if len(started) != 0 {
		t.Errorf("GET /setup provisioned the lab")
	}
}

func TestSetupProvisionsTheLabOnce(t *testing.T) {
	started, release := stubProvision(t)

	if rec := requestSetup(http.MethodPost); rec.Code != http.StatusAccepted {
		t.Fatalf("first POST: status %d, want 202", rec.Code)
	}
	<-started // provisioning is now running

	if rec := requestSetup(http.MethodPost); rec.Code != http.StatusConflict {
		t.Errorf("second POST while running: status %d, want 409", rec.Code)
	}
	if len(started) != 0 {
		t.Errorf("a second provisioning started while the first was running")
	}

	// Once it finishes, setup can be requested again.
	close(release)
	deadline := time.Now().Add(2 * time.Second)
	for setupRunning.Load() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if setupRunning.Load() {
		t.Fatal("setup still marked as running after it finished")
	}
}

// The interface served by this binary must be able to call the API and open
// the terminal, whichever loopback name the browser used.
func TestOwnOriginIsAlwaysAllowed(t *testing.T) {
	for _, configured := range []string{"", "http://example.test"} {
		origins := parseAllowedOrigins(configured)
		for _, own := range []string{"http://127.0.0.1:8083", "http://localhost:8083"} {
			if !origins[own] {
				t.Errorf("with KUBELEARN_ALLOWED_ORIGINS=%q, %s is not allowed", configured, own)
			}
		}
	}
	if parseAllowedOrigins("")["http://evil.example:8083"] {
		t.Errorf("a foreign host on the same port is allowed")
	}
}

func TestSPAHandler(t *testing.T) {
	files := fstest.MapFS{
		"index.html":        {Data: []byte("<html>app</html>")},
		"static/js/main.js": {Data: []byte("console.log(1)")},
		"images/logo.png":   {Data: []byte("png")},
	}
	handler := spaHandler(files)

	cases := []struct{ path, want string }{
		{"/", "<html>app</html>"},
		{"/static/js/main.js", "console.log(1)"},
		{"/images/logo.png", "png"},
		{"/some/client/route", "<html>app</html>"}, // unknown paths get the app
		{"/static", "<html>app</html>"},            // a directory is not a file
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != tc.want {
			t.Errorf("%s: status %d body %q, want %q", tc.path, rec.Code, rec.Body.String(), tc.want)
		}
	}
}

// API routes keep working next to the interface, and a build without the
// interface says how to get one.
func TestMuxServesAPIAndInterface(t *testing.T) {
	mux := newMux(emptyCluster())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/catalog", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("/catalog: status %d, content type %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if _, embedded := kubelearn.Frontend(); embedded {
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<div id=\"root\">") {
			t.Errorf("/: the embedded interface was not served (status %d)", rec.Code)
		}
	} else if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "make build") {
		t.Errorf("/: status %d body %q, want the build instructions", rec.Code, rec.Body.String())
	}
}

func TestQuestionsEndpointListsEveryQuestion(t *testing.T) {
	rec := httptest.NewRecorder()
	getQuestions(rec, httptest.NewRequest(http.MethodGet, "/questions", nil), emptyCluster())

	var results []utils.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(results) != len(resources.Questions) {
		t.Errorf("got %d questions, want %d", len(results), len(resources.Questions))
	}
}

func TestFinishEndpointReturnsScoreAndResults(t *testing.T) {
	rec := httptest.NewRecorder()
	finishQuiz(rec, httptest.NewRequest(http.MethodGet, "/finish", nil), emptyCluster())

	var body struct {
		Score   *float64       `json:"score"`
		Results []utils.Result `json:"results"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.Score == nil || *body.Score != 0 {
		t.Errorf("score on an empty cluster = %v, want 0", body.Score)
	}
	if len(body.Results) != len(resources.Questions) {
		t.Errorf("got %d results, want %d", len(body.Results), len(resources.Questions))
	}
}

// When the cluster does not answer in time, the endpoints must report it
// instead of hanging or returning every question as failed.
func TestEvaluationTimeoutReturnsGatewayTimeout(t *testing.T) {
	original := evaluationTimeout
	evaluationTimeout = time.Nanosecond
	t.Cleanup(func() { evaluationTimeout = original })

	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/questions": func(w http.ResponseWriter, r *http.Request) { getQuestions(w, r, emptyCluster()) },
		"/finish":    func(w http.ResponseWriter, r *http.Request) { finishQuiz(w, r, emptyCluster()) },
	}
	for path, handler := range handlers {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusGatewayTimeout {
			t.Errorf("%s: status %d, want 504", path, rec.Code)
		}
	}
}

func TestEvaluationStopsWhenClientDisconnects(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/questions", nil).WithContext(ctx)

	rec := httptest.NewRecorder()
	getQuestions(rec, req, emptyCluster())
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", rec.Code)
	}
}

// A failing command must still show what it printed, not only the exit status.
func TestTerminalCommandKeepsOutputOnFailure(t *testing.T) {
	got := string(runTerminalCommand("echo partial output; echo the real reason >&2; exit 3"))
	for _, want := range []string{"partial output", "the real reason", "[exit status 3]"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q is missing %q", got, want)
		}
	}
	if strings.Index(got, "the real reason") > strings.Index(got, "[exit status 3]") {
		t.Errorf("exit status should come after the command output: %q", got)
	}
}

func TestTerminalCommandSuccessOutputIsUnchanged(t *testing.T) {
	if got := string(runTerminalCommand("echo ok")); got != "ok\n" {
		t.Errorf("output = %q, want %q", got, "ok\n")
	}
}

func emptyCluster() resources.Clients {
	return resources.Clients{
		Kube:    fake.NewSimpleClientset(),
		Dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
	}
}

func TestQuestionsCanBeLimitedToATrack(t *testing.T) {
	for query, want := range map[string]int{"": len(resources.Questions), "?track=all": len(resources.Questions), "?track=cks": 11, "?track=CKNE": 17, "?track=core": 26} {
		rec := httptest.NewRecorder()
		getQuestions(rec, httptest.NewRequest(http.MethodGet, "/questions"+query, nil), emptyCluster())

		var results []utils.Result
		if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
			t.Fatalf("%q: invalid JSON: %v", query, err)
		}
		if len(results) != want {
			t.Errorf("%q: got %d questions, want %d", query, len(results), want)
		}
	}
}

func TestUnknownTrackIsRejected(t *testing.T) {
	rec := httptest.NewRecorder()
	finishQuiz(rec, httptest.NewRequest(http.MethodGet, "/finish?track=cka", nil), emptyCluster())
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}

func TestCatalogEndpointNeedsNoCluster(t *testing.T) {
	rec := httptest.NewRecorder()
	getCatalog(rec, httptest.NewRequest(http.MethodGet, "/catalog", nil))

	var entries []utils.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != len(resources.Questions) {
		t.Errorf("got %d entries, want %d", len(entries), len(resources.Questions))
	}
	if entries[len(entries)-1].Track != resources.TrackCKNE {
		t.Errorf("last entry track = %q, want CKNE", entries[len(entries)-1].Track)
	}
}

func TestQuestionsCanBeLimitedToADifficulty(t *testing.T) {
	rec := httptest.NewRecorder()
	getQuestions(rec, httptest.NewRequest(http.MethodGet, "/questions?track=cks&difficulty=hard", nil), emptyCluster())

	var results []utils.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("got %d questions for CKS + hard, want 4", len(results))
	}

	rec = httptest.NewRecorder()
	getQuestions(rec, httptest.NewRequest(http.MethodGet, "/questions?difficulty=extreme", nil), emptyCluster())
	if rec.Code != http.StatusBadRequest {
		t.Errorf("unknown difficulty: status %d, want 400", rec.Code)
	}
}

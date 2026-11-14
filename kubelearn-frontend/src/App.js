import React, { useEffect, useMemo, useState, useId } from 'react';
import TerminalPanel from './components/TerminalPanel';

const PASS_THRESHOLD = 85;

const BASE_QUESTIONS = [
  { difficulty: 'Easy', description: 'Create a new pod called `nginx` with `nginx:alpine` image in `default` namespace.' },
  { difficulty: 'Medium', description: 'Create a new deployment called `nginx-deployment` with `nginx:alpine` image and `4 replicas` in default namespace.' },
  { difficulty: 'Hard', description: 'Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace, and create a service called `redis-service` with port `6379` in same namespace.' },
  { difficulty: 'Easy', description: 'Create a namespace called `europe`.' },
  { difficulty: 'Medium', description: 'Create a configmap `europe-configmap` with data `France=Paris`.' },
  { difficulty: 'Medium', description: 'Create a pod `tshoot` with label `country=china`, image `amazon/amazon-ecs-network-sidecar:latest`, in namespace `asia`.' },
  { difficulty: 'Medium', description: 'Create a persistent volume `unicorn-pv` with capacity `1Gi`, access mode `ReadWriteMany`, and host path `/tmp/data`.' },
  { difficulty: 'Medium', description: 'Create a persistent volume claim `unicorn-pvc` with capacity `400Mi` and access mode `ReadWriteMany`.' },
  { difficulty: 'Hard', description: 'Create a pod `webserver` in namespace `public` with `nginx:alpine`, mount `/usr/share/nginx/html`, and use PVC `unicorn-pvc`.' },
  { difficulty: 'Hard', description: 'Troubleshoot a failing pod and ensure it runs successfully.' },
  { difficulty: 'Hard', description: 'Create NetworkPolicy `allow-policy-colors` allowing `redmobile-webserver` to reach `bluemobile-dbcache` in `colors` namespace.' },
  { difficulty: 'Easy', description: 'Create secret `secret-colors` with data `color=red` in namespace `colors`.' },
  { difficulty: 'Hard', description: 'Add secret `secret-purple` with data `singer=prince` to pod `purple` (image `redis:alpine`) in namespace `colors`.' },
  { difficulty: 'Easy', description: 'Create a service account `america-sa` in namespace `default`.' },
  { difficulty: 'Medium', description: 'Assign service account `america-sa` to deployment `mark42`.' },
  { difficulty: 'Medium', description: 'Scale deployment `mark42` to `5` replicas.' },
  { difficulty: 'Medium', description: 'Create HPA for deployment `mark43` (CPU 80%, min 2 replicas, max 8 replicas).' },
  { difficulty: 'Medium', description: 'Prevent privilege escalation in deployment `mark42`.' },
  { difficulty: 'Medium', description: 'Add a liveness probe (delay 5s, period 10s, path `/`) to pod `mark50` in namespace `shield`.' },
  { difficulty: 'Easy', description: 'Create deployment `yellow-deployment` with `bonovoo/node-app:1.0` image and 2 replicas in namespace `colors`.' },
  { difficulty: 'Hard', description: 'Expose deployment `yellow-deployment` with service `yellow-service` (port 80 -> targetPort 3000) in namespace `colors`.' },
  { difficulty: 'Hard', description: 'Create ingress `ingress-colors` for host `yellow.com`, path `/yellow`, targeting service `yellow-service` in namespace `colors`.' },
  { difficulty: 'Hard', description: 'Create role `apple-one` with verbs `get, list, watch` in namespace `fruits`.' },
  { difficulty: 'Medium', description: 'Create job `job-gain` with parallelism `2`, completions `4`, backoffLimit `3`, activeDeadlineSeconds `40`.' },
  { difficulty: 'Medium', description: 'Create cronjob `cronjob-gain` every 5 minutes running `busybox:1.28` with command `sleep 3600`, restartPolicy `Never`.' },
  { difficulty: 'Hard', description: 'Create statefulset `statefulset-gain` with image `busybox:1.28`, command `sleep 3600`, replicas `3`.' },
];

const ADVANCED_CKS_QUESTIONS = [
  { difficulty: 'Hard', description: 'Create pod that mounts projected volume with TLS cert/key secrets and exports them as env vars.' },
  { difficulty: 'Hard', description: 'Rotate secret `db-credentials` used by deployment `inventory-api` and roll pods to consume new values.' },
  { difficulty: 'Hard', description: 'Enable audit policy in namespace `audit-labs` to log Secret GET operations and produce one audited request.' },
  { difficulty: 'Hard', description: 'Simulate encrypted storage: create CSI secret-backed volume for pod `vault-agent` and prove secret never appears in plain text on node.' },
  { difficulty: 'Hard', description: 'Author CronJob `etcd-backup` that runs daily, stores etcd snapshot to PVC, and deletes files older than 7 days.' },
  { difficulty: 'Hard', description: 'Deploy DaemonSet `falco-agent` watching privileged activity; trigger an alert by `exec` into a privileged pod and note event.' },
  { difficulty: 'Hard', description: 'Create pod `image-policy-test` that downloads from private registry using image policy webhook credentials stored in secret `policy-creds`.' },
];

const ADVANCED_QUESTIONS = ADVANCED_CKS_QUESTIONS.map((item, idx) => ({
  id: BASE_QUESTIONS.length + idx + 1,
  TestName: `Advanced ${idx + 1} - ${item.description}`,
  Difficulty: item.difficulty,
  Advanced: true,
  Passed: null,
}));

const FALLBACK_QUESTIONS = BASE_QUESTIONS.map((item, idx) => ({
  id: idx + 1,
  TestName: `Question ${idx + 1} - ${item.description}`,
  Difficulty: item.difficulty,
  Track: 'Core',
  Advanced: false,
  Passed: null,
}));

const API = 'http://localhost:8083';

// Exam tracks offered on the start screen. "all" runs every question.
const TRACKS = [
  { id: 'all', label: 'All', about: 'Every question' },
  { id: 'Core', label: 'Core', about: 'General Kubernetes practice' },
  { id: 'CKS', label: 'CKS', about: 'Security' },
  { id: 'CKNE', label: 'CKNE', about: 'Networking' },
];

const inTrack = (track) => (question) => track === 'all' || question.Track === track;

const difficultyColor = {
  easy: 'text-kube-soft',
  medium: 'text-warn',
  hard: 'text-fail',
};

const focusRing =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-kube-soft';

const primaryButtonClasses = `inline-flex items-center justify-center rounded-md bg-kube px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-[#4a7ff0] disabled:cursor-not-allowed disabled:opacity-50 ${focusRing}`;

const secondaryButtonClasses = `inline-flex items-center justify-center rounded-md border border-line px-4 py-2 text-sm font-semibold text-fg transition-colors hover:border-kube-soft disabled:cursor-not-allowed disabled:opacity-50 ${focusRing}`;

const quietButtonClasses = `rounded-md px-2 py-2 text-sm text-muted transition-colors hover:text-fg ${focusRing}`;

const extractObjective = (testName = '') => {
  if (!testName) return 'Question not available.';
  const stripped = testName.replace(/^(Question|Challenge|Advanced)\s*\d+\s*[-–—:]\s*/i, '').trim();
  return stripped.length ? stripped : testName;
};

// Renders `names` wrapped in backticks as inline code.
const renderInline = (text) =>
  text.split('`').map((part, idx) =>
    idx % 2 === 1 ? (
      <code key={idx} className="rounded bg-panel px-1.5 py-0.5 font-mono text-[0.9em] text-kube-soft">
        {part}
      </code>
    ) : (
      part
    ),
  );

const formatTime = (seconds) => {
  const minutes = Math.floor(seconds / 60)
    .toString()
    .padStart(2, '0');
  const remainingSeconds = (seconds % 60).toString().padStart(2, '0');
  return `${minutes}:${remainingSeconds}`;
};

// One question as a list row. The left edge shows its state once it has been
// checked: green when solved, red when not.
const QuestionRow = ({ question, number, showResult, showTrack = false, failedLabel = 'Review needed' }) => {
  const objective = extractObjective(question.TestName);
  const hasResult = showResult && typeof question.Passed === 'boolean';
  const titleId = useId();
  const edge = !hasResult ? 'border-l-transparent' : question.Passed ? 'border-l-pass' : 'border-l-fail';

  return (
    <li
      aria-labelledby={titleId}
      className={`grid break-inside-avoid grid-cols-[2rem_minmax(0,1fr)] gap-x-3 border-b border-l-2 border-b-line px-5 py-4 ${edge}`}
    >
      <span className="pt-0.5 font-mono text-sm tabular-nums text-muted" aria-hidden="true">
        {number}
      </span>
      <div>
        <h3 id={titleId} className="break-words text-[0.95rem] font-normal leading-relaxed text-fg">
          {renderInline(objective)}
        </h3>
        <p className="mt-1.5 flex flex-wrap items-baseline gap-x-3 text-sm">
          <span className={difficultyColor[question.Difficulty?.toLowerCase?.()] || 'text-muted'}>
            {question.Difficulty || 'Unknown'}
          </span>
          {showTrack && question.Track && <span className="text-muted">{question.Track}</span>}
          {question.Advanced && <span className="text-muted">Extra scenario, not graded</span>}
          {hasResult && (
            <span className={`font-semibold ${question.Passed ? 'text-pass' : 'text-fail'}`}>
              {question.Passed ? 'Passed' : failedLabel}
            </span>
          )}
        </p>
        {hasResult && !question.Passed && question.Feedback && (
          <p className="mt-1.5 text-sm leading-relaxed text-muted">{question.Feedback}</p>
        )}
      </div>
    </li>
  );
};

const ErrorBanner = ({ message }) =>
  message ? (
    <div className="border-b border-fail/40 bg-fail/10 px-5 py-3 text-sm text-fail" role="alert" aria-live="polite">
      {message}
    </div>
  ) : null;

function App() {
  const [questions, setQuestions] = useState(FALLBACK_QUESTIONS);
  const [isFallback, setIsFallback] = useState(true);
  const [includeAdvanced, setIncludeAdvanced] = useState(false);
  const [quizStarted, setQuizStarted] = useState(false);
  const [quizFinished, setQuizFinished] = useState(false);
  const [score, setScore] = useState(0);
  const [elapsedTime, setElapsedTime] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [isFinishing, setIsFinishing] = useState(false);
  const [error, setError] = useState('');
  const [lastAttempt, setLastAttempt] = useState(null);
  const [isChecking, setIsChecking] = useState(false);
  const [lastCheckedAt, setLastCheckedAt] = useState(null);
  const [track, setTrack] = useState('all');
  // Every question as described by the backend; null until it answers.
  const [catalog, setCatalog] = useState(null);

  const trackQuery = track === 'all' ? '' : `?track=${track.toLowerCase()}`;

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const response = await fetch(`${API}/catalog`);
        if (!response.ok) return;
        const data = await response.json();
        if (!cancelled && Array.isArray(data) && data.length > 0) setCatalog(data);
      } catch (err) {
        // Backend not running: the start screen falls back to the built-in list.
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const fetchQuestions = async () => {
    setIsLoading(true);
    setError('');
    try {
      const response = await fetch(`${API}/questions${trackQuery}`);
      if (!response.ok) throw new Error('Failed to fetch questions');
      const data = await response.json();
      if (Array.isArray(data) && data.length > 0) {
        setQuestions(data);
        setIsFallback(false);
      } else {
        setQuestions(FALLBACK_QUESTIONS.filter(inTrack(track)));
        setIsFallback(true);
      }
    } catch (err) {
      const reason = err instanceof Error ? err.message : 'Unknown error';
      console.warn(`[KubeLearn] Unable to fetch questions: ${reason}`);
      setError('Unable to sync the questions. Make sure the backend is running and try again.');
      setQuestions(FALLBACK_QUESTIONS.filter(inTrack(track)));
      setIsFallback(true);
    } finally {
      setIsLoading(false);
    }
  };

  // Re-evaluates the cluster mid-attempt so each challenge shows whether it is
  // solved and, if not, why. The list stays on screen while the check runs.
  const checkProgress = async () => {
    if (isChecking) return;
    setIsChecking(true);
    setError('');
    try {
      const response = await fetch(`${API}/questions${trackQuery}`);
      if (!response.ok) throw new Error(`Backend answered ${response.status}`);
      const data = await response.json();
      if (!Array.isArray(data) || data.length === 0) throw new Error('No questions returned');
      setQuestions(data);
      setIsFallback(false);
      setLastCheckedAt(new Date());
    } catch (err) {
      const reason = err instanceof Error ? err.message : 'Unknown error';
      console.warn(`[KubeLearn] Unable to check progress: ${reason}`);
      setError('Unable to check your progress. Make sure the backend and the cluster are running and try again.');
    } finally {
      setIsChecking(false);
    }
  };

  const startQuiz = () => {
    setLastCheckedAt(null);
    setQuizStarted(true);
    setQuizFinished(false);
    setScore(0);
    setElapsedTime(0);
    setError('');
    fetchQuestions();
  };

  const finishQuiz = async () => {
    if (isFinishing) return;
    setIsFinishing(true);
    setError('');
    try {
      const response = await fetch(`${API}/finish${trackQuery}`);
      if (!response.ok) throw new Error('Failed to finish quiz');
      const data = await response.json();
      const normalizedScore = Math.round(data.score);
      // Show the results the score was computed from, not the snapshot taken
      // when the quiz started.
      const hasResults = Array.isArray(data.results) && data.results.length > 0;
      const results = hasResults ? data.results : questions;
      if (hasResults) {
        setQuestions(results);
        setIsFallback(false);
      }
      setScore(normalizedScore);
      setQuizFinished(true);

      const passedQuestions = results.filter((q) => q.Passed).length;
      setLastAttempt({
        score: normalizedScore,
        totalQuestions: results.length,
        passedQuestions,
        duration: elapsedTime,
        track,
        completedAt: new Date().toISOString(),
      });
    } catch (err) {
      const reason = err instanceof Error ? err.message : 'Unknown error';
      console.warn(`[KubeLearn] Unable to finish quiz: ${reason}`);
      setError('Unable to finish the quiz. Confirm the backend is running and try again.');
    } finally {
      setIsFinishing(false);
    }
  };

  const resetQuiz = () => {
    setLastCheckedAt(null);
    setQuizStarted(false);
    setQuizFinished(false);
    setQuestions(FALLBACK_QUESTIONS);
    setIsFallback(true);
    setScore(0);
    setElapsedTime(0);
    setError('');
    setIsLoading(false);
    setIsFinishing(false);
  };

  useEffect(() => {
    let timer;
    if (quizStarted && !quizFinished) {
      timer = setInterval(() => {
        setElapsedTime((prev) => prev + 1);
      }, 1000);
    }
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [quizStarted, quizFinished]);

  // Before an attempt the list is a preview of the chosen track; during and
  // after it, the questions the backend evaluated.
  const preview = (catalog ?? FALLBACK_QUESTIONS).filter(inTrack(track));
  const evaluatedQuestions = quizStarted ? questions : preview;
  const trackCount = (id) => (catalog ?? FALLBACK_QUESTIONS).filter(inTrack(id)).length;
  const trackLabel = TRACKS.find((item) => item.id === track)?.label ?? track;
  const showExtras = track === 'all' || track === 'CKS';
  const displayedQuestions = useMemo(
    () => (includeAdvanced && showExtras ? [...evaluatedQuestions, ...ADVANCED_QUESTIONS] : evaluatedQuestions),
    [evaluatedQuestions, includeAdvanced, showExtras],
  );

  const evaluatedCount = evaluatedQuestions.length;
  const passedCount = evaluatedQuestions.filter((q) => q.Passed).length;
  const progressPercent = evaluatedCount ? Math.round((passedCount / evaluatedCount) * 100) : 0;
  const numbered = evaluatedQuestions.map((question, index) => ({ question, number: index + 1 }));

  // Start screen: what this is, one action, and the real question list.
  if (!quizStarted) {
    return (
      <main className="mx-auto max-w-5xl px-6 py-14 sm:py-20">
        <h1 className="text-4xl font-semibold tracking-tight text-fg sm:text-5xl">KubeLearn</h1>
        <p className="mt-4 max-w-2xl text-lg leading-relaxed text-muted">
          Hands-on Kubernetes questions. Solve each one in your cluster from the built-in terminal, then check your
          answers against the live cluster.
        </p>

        <fieldset className="mt-10">
          <legend className="text-sm text-muted">Choose what to practise</legend>
          <div className="mt-3 flex flex-wrap gap-3">
            {TRACKS.map((item) => {
              const selected = item.id === track;
              const count = trackCount(item.id);
              return (
                <label
                  key={item.id}
                  className={`cursor-pointer rounded-md border px-4 py-3 transition-colors focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-kube-soft ${
                    selected ? 'border-kube bg-kube/15' : 'border-line hover:border-kube-soft'
                  }`}
                >
                  <input
                    type="radio"
                    name="track"
                    value={item.id}
                    checked={selected}
                    onChange={() => setTrack(item.id)}
                    className="sr-only"
                    aria-label={`${item.label} track, ${count} questions`}
                  />
                  <span className="block text-base font-semibold text-fg">
                    {item.label} <span className="font-normal tabular-nums text-muted">{count}</span>
                  </span>
                  <span className="block text-sm text-muted">{item.about}</span>
                </label>
              );
            })}
          </div>
          {!catalog && (
            <p className="mt-3 text-sm text-warn">
              The backend is not answering, so only the built-in Core list is shown. Start it with make backend-start.
            </p>
          )}
        </fieldset>

        <div className="mt-8 flex flex-wrap items-center gap-x-6 gap-y-3">
          <button
            type="button"
            onClick={startQuiz}
            className={primaryButtonClasses}
            aria-label="Start quiz simulation"
            disabled={evaluatedCount === 0}
          >
            Start attempt with {evaluatedCount} questions
          </button>
          {lastAttempt && (
            <p className="text-sm text-muted">
              Last attempt: {lastAttempt.score}% ({lastAttempt.passedQuestions} of {lastAttempt.totalQuestions}) in{' '}
              {formatTime(lastAttempt.duration)}, {new Date(lastAttempt.completedAt).toLocaleString()}
            </p>
          )}
        </div>

        <h2 className="mt-16 text-lg font-semibold text-fg">The questions</h2>
        <ol className="mt-4 border-t border-line lg:columns-2 lg:gap-x-10">
          {numbered.map(({ question, number }) => (
            <QuestionRow
              key={question.TestName}
              question={question}
              number={number}
              showResult={false}
              showTrack={track === 'all'}
            />
          ))}
        </ol>
      </main>
    );
  }

  // Results: the score first, then what is left to solve and why.
  if (quizFinished) {
    const passedOverall = score >= PASS_THRESHOLD;
    const unsolved = numbered.filter(({ question }) => !question.Passed);
    const solved = numbered.filter(({ question }) => question.Passed);

    return (
      <main className="mx-auto max-w-5xl px-6 py-14">
        <h1 className="text-lg font-semibold text-muted">Results overview</h1>
        <div className="mt-4 flex flex-wrap items-end gap-x-10 gap-y-4">
          <p className={`text-7xl font-semibold leading-none tabular-nums ${passedOverall ? 'text-pass' : 'text-fg'}`}>
            {score}%
          </p>
          <div className="space-y-1 text-base">
            <p className={`font-semibold ${passedOverall ? 'text-pass' : 'text-warn'}`}>
              {passedOverall ? 'Status: Passed' : 'Status: Keep practising'}
            </p>
            <p className="text-fg">
              {passedCount} of {evaluatedCount} questions solved.
            </p>
            <p className="text-muted">
              Time spent {formatTime(elapsedTime)}. The passing score is {PASS_THRESHOLD}%.
            </p>
          </div>
        </div>
        <div className="mt-8 flex flex-wrap items-center gap-3">
          <button type="button" onClick={startQuiz} className={primaryButtonClasses}>
            Start another attempt
          </button>
          <button type="button" onClick={resetQuiz} className={secondaryButtonClasses}>
            Back to start
          </button>
        </div>

        {unsolved.length > 0 && (
          <section className="mt-14">
            <h2 className="text-lg font-semibold text-fg">Still to solve ({unsolved.length})</h2>
            <ol className="mt-4 border-t border-line">
              {unsolved.map(({ question, number }) => (
                <QuestionRow key={question.TestName} question={question} number={number} showResult />
              ))}
            </ol>
          </section>
        )}
        {solved.length > 0 && (
          <section className="mt-14">
            <h2 className="text-lg font-semibold text-fg">Solved ({solved.length})</h2>
            <ol className="mt-4 border-t border-line">
              {solved.map(({ question, number }) => (
                <QuestionRow key={question.TestName} question={question} number={number} showResult />
              ))}
            </ol>
          </section>
        )}
      </main>
    );
  }

  // Attempt in progress: a workbench that fills the window. Questions on the
  // left, terminal on the right, each scrolling on its own.
  return (
    <div className="flex min-h-screen flex-col lg:h-screen">
      <header className="flex flex-wrap items-center gap-x-6 gap-y-3 border-b border-line bg-panel px-5 py-3">
        <p className="text-base font-semibold text-fg">
          KubeLearn{track !== 'all' && <span className="font-normal text-muted"> {trackLabel}</span>}
        </p>
        <p className="font-mono text-base tabular-nums text-fg" role="timer" aria-label="Time elapsed">
          {formatTime(elapsedTime)}
        </p>
        <div className="flex min-w-[14rem] flex-1 items-center gap-3">
          <div className="h-1.5 w-32 shrink-0 overflow-hidden rounded-full bg-line" aria-hidden="true">
            <div className="h-full rounded-full bg-pass transition-[width] duration-300" style={{ width: `${lastCheckedAt ? progressPercent : 0}%` }} />
          </div>
          <p className="text-sm text-muted" role="status" aria-live="polite">
            {lastCheckedAt
              ? `${passedCount} of ${evaluatedCount} solved, checked at ${lastCheckedAt.toLocaleTimeString()}`
              : isFallback && !isLoading
                ? 'Showing the built-in question list'
                : `${evaluatedCount} questions, not checked yet`}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={checkProgress}
            className={secondaryButtonClasses}
            disabled={isChecking || isFinishing || isLoading}
            aria-label="Check progress on the challenges"
          >
            {isChecking ? 'Checking…' : 'Check progress'}
          </button>
          <button
            type="button"
            onClick={finishQuiz}
            className={primaryButtonClasses}
            disabled={isFinishing || isLoading}
            aria-label="Finish current attempt"
          >
            {isFinishing ? 'Finishing…' : 'Finish attempt'}
          </button>
          <button type="button" onClick={startQuiz} className={quietButtonClasses} disabled={isLoading} aria-label="Restart quiz simulation">
            Restart
          </button>
          <button type="button" onClick={resetQuiz} className={quietButtonClasses} aria-label="Reset quiz session">
            Leave
          </button>
        </div>
      </header>

      <ErrorBanner message={error} />

      <div className="grid min-h-0 flex-1 lg:grid-cols-[minmax(0,5fr)_minmax(0,7fr)]">
        <section className="min-h-0 border-line lg:overflow-y-auto lg:border-r" aria-label="Questions">
          {isLoading ? (
            <p className="px-5 py-6 text-sm text-muted">Loading questions…</p>
          ) : (
            <ol aria-label="Kubernetes challenges">
              {displayedQuestions.map((question, index) => (
                <QuestionRow
                  key={`${question.TestName}-${index}`}
                  question={question}
                  number={index + 1}
                  showResult={lastCheckedAt !== null}
                  showTrack={track === 'all'}
                  failedLabel="Not solved yet"
                />
              ))}
            </ol>
          )}
          {showExtras && (
            <div className="px-5 py-4">
              <button
                type="button"
                onClick={() => setIncludeAdvanced((prev) => !prev)}
                className={quietButtonClasses}
                aria-pressed={includeAdvanced}
              >
                {includeAdvanced ? 'Hide extra CKS scenarios' : 'Show extra CKS scenarios (not graded)'}
              </button>
            </div>
          )}
        </section>

        <section className="h-[70vh] min-h-0 border-t border-line lg:h-auto lg:border-t-0" aria-label="Terminal">
          <TerminalPanel isVisible />
        </section>
      </div>
    </div>
  );
}

export default App;

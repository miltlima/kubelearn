import React, { useEffect, useMemo, useState, useId } from 'react';
import TerminalPanel from './components/TerminalPanel';

const PASS_THRESHOLD = 85;

const BASE_QUESTIONS = [
  { difficulty: 'Easy', description: 'Create a new pod called `nginx` with `nginx:alpine` image in `default` namespace.' },
  { difficulty: 'Medium', description: 'Create a new deployment called `nginx-deployment` with `nginx:alpine` image and `4 replicas` in default namespace.' },
  { difficulty: 'Hard', description: 'Create a new deployment called `redis` with image `redis:alpine` in `latam` namespace, and create a service called `redis-service` with port `6379` in same namespace.' },
  { difficulty: 'Easy', description: 'Create a namespace called `europe`.' },
  { difficulty: 'Medium', description: 'Create a configmap `europe-configmap` with data `France=Paris`.' },
  { difficulty: 'Medium', description: 'Create a pod `thsoot` with label `country=china`, image `amazon/amazon-ecs-network-sidecar:latest`, in namespace `asia`.' },
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
  { difficulty: 'Hard', description: 'Run pod `nginx-apparmor` applying AppArmor profile `localhost/kubelearn-nginx`.' },
  { difficulty: 'Hard', description: 'Launch pod `seccomp-default` with seccomp profile `runtime/default` and document how to verify it.' },
  { difficulty: 'Hard', description: 'Create pod that mounts projected volume with TLS cert/key secrets and exports them as env vars.' },
  { difficulty: 'Hard', description: 'Rotate secret `db-credentials` used by deployment `inventory-api` and roll pods to consume new values.' },
  { difficulty: 'Hard', description: 'Enable audit policy in namespace `audit-labs` to log Secret GET operations and produce one audited request.' },
  { difficulty: 'Hard', description: 'Deploy pod `strict-security` forcing non-root UID, dropping all capabilities, adding only `NET_BIND_SERVICE`, and mounting root filesystem read-only.' },
  { difficulty: 'Hard', description: 'Create NetworkPolicy in namespace `secure-egress` that denies all egress except UDP/53 to cluster DNS and TCP to CIDR `10.10.0.0/16`.' },
  { difficulty: 'Hard', description: 'Simulate encrypted storage: create CSI secret-backed volume for pod `vault-agent` and prove secret never appears in plain text on node.' },
  { difficulty: 'Hard', description: 'Author CronJob `etcd-backup` that runs daily, stores etcd snapshot to PVC, and deletes files older than 7 days.' },
  { difficulty: 'Hard', description: 'Deploy DaemonSet `falco-agent` watching privileged activity; trigger an alert by `exec` into a privileged pod and note event.' },
  { difficulty: 'Hard', description: 'Configure PodSecurity `baseline` for namespace `baseline-labs` and `restricted` for `restricted-labs`, then validate compliance.' },
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
  Advanced: false,
  Passed: null,
}));

const difficultyPalette = {
  easy: 'bg-blue-100 text-blue-700',
  medium: 'bg-amber-100 text-amber-700',
  hard: 'bg-rose-100 text-rose-700',
};

const primaryButtonClasses =
  'inline-flex items-center justify-center gap-2 rounded-full bg-blue-600 px-6 py-3 text-base font-semibold text-white transition duration-200 hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400 disabled:cursor-not-allowed disabled:bg-blue-300';

const secondaryButtonClasses =
  'inline-flex items-center justify-center gap-2 rounded-full border border-slate-200 bg-white px-6 py-3 text-base font-semibold text-slate-700 transition duration-200 hover:border-blue-300 hover:text-blue-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400 disabled:cursor-not-allowed disabled:opacity-50';

const StatCard = ({ label, value, helper, accent }) => (
  <div className="rounded-[24px] border border-slate-200 bg-white px-6 py-5 shadow-[0_14px_32px_rgba(15,23,42,0.08)]">
    <p className="text-xs font-semibold uppercase tracking-[0.32em] text-slate-500">{label}</p>
    <p className={`mt-3 text-3xl font-semibold text-slate-900 ${accent ?? ''}`}>{value}</p>
    {helper && <p className="mt-2 text-base leading-relaxed text-slate-500">{helper}</p>}
  </div>
);

const DifficultyBadge = ({ value }) => {
  const cls = difficultyPalette[value?.toLowerCase?.()] || 'bg-slate-200 text-slate-700';
  return <span className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold ${cls}`}>{value || 'Unknown'}</span>;
};

const extractObjective = (testName = '') => {
  if (!testName) return 'Challenge objective not available.';
  const stripped = testName.replace(/^(Question|Challenge)\s*\d+\s*[-–—:]\s*/i, '').trim();
  return stripped.length ? stripped : testName;
};

const QuestionCard = ({ question, index, showResult }) => {
  const objective = extractObjective(question.TestName);
  const hasResult = showResult && typeof question.Passed === 'boolean';
  const bulletPoints = objective
    .split(/\.(?=\s|$)/)
    .map((part) => part.trim())
    .filter(Boolean);
  const titleId = useId();

  return (
    <article
      role="listitem"
      aria-labelledby={titleId}
      className="rounded-2xl border border-slate-200 bg-white/95 p-6 shadow-sm focus-within:border-blue-300 focus-within:shadow-[0_0_0_2px_rgba(37,99,235,0.2)]"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <p className="text-sm font-semibold uppercase tracking-[0.35em] text-slate-500">Challenge {index + 1}</p>
          {question.Advanced && (
            <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-0.5 text-[0.65rem] font-medium uppercase tracking-[0.3em] text-slate-500">
              Advanced
            </span>
          )}
        </div>
        <DifficultyBadge value={question.Difficulty} />
      </div>
      <p className="mt-3 text-sm font-semibold uppercase tracking-[0.28em] text-slate-400">Objective</p>
      <h3 id={titleId} className="sr-only">
        {objective}
      </h3>

      {bulletPoints.length > 1 ? (
        <ul className="mt-2 space-y-1.5 text-lg leading-relaxed text-slate-800">
          {bulletPoints.map((point, idx) => (
            <li key={idx} className="flex gap-2">
              <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-blue-500" />
              <span className="whitespace-pre-line break-words">{point}</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="mt-2 text-lg leading-relaxed text-slate-800 whitespace-pre-line break-words">{objective}</p>
      )}

      {hasResult && (
        <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
          <span
            className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold ${
              question.Passed
                ? 'bg-blue-100 text-blue-700 ring-1 ring-inset ring-blue-200'
                : 'bg-rose-100 text-rose-700 ring-1 ring-inset ring-rose-200'
            }`}
          >
            {question.Passed ? 'Passed' : 'Review needed'}
          </span>
          {question.Feedback && <p className="text-slate-500">{question.Feedback}</p>}
        </div>
      )}
    </article>
  );
};

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

  const fetchQuestions = async () => {
    setIsLoading(true);
    setError('');
    try {
      const response = await fetch('http://localhost:8083/questions');
      if (!response.ok) throw new Error('Failed to fetch questions');
      const data = await response.json();
      if (Array.isArray(data) && data.length > 0) {
        setQuestions(data);
        setIsFallback(false);
      } else {
        setQuestions(FALLBACK_QUESTIONS);
        setIsFallback(true);
      }
    } catch (err) {
      const reason = err instanceof Error ? err.message : 'Unknown error';
      console.warn(`[KubeLearn] Unable to fetch questions: ${reason}`);
      setError('Unable to sync the questions. Make sure the backend is running and try again.');
      setQuestions(FALLBACK_QUESTIONS);
      setIsFallback(true);
    } finally {
      setIsLoading(false);
    }
  };

  const startQuiz = () => {
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
      const response = await fetch('http://localhost:8083/finish');
      if (!response.ok) throw new Error('Failed to finish quiz');
      const data = await response.json();
      const normalizedScore = Math.round(data.score);
      setScore(normalizedScore);
      setQuizFinished(true);

      const passedQuestions = questions.filter((q) => q.Passed).length;
      setLastAttempt({
        score: normalizedScore,
        totalQuestions: questions.length,
        passedQuestions,
        duration: elapsedTime,
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

  const evaluatedQuestions = questions.length ? questions : FALLBACK_QUESTIONS;
  const displayedQuestions = useMemo(
    () => (includeAdvanced ? [...evaluatedQuestions, ...ADVANCED_QUESTIONS.map((q) => ({ ...q, Passed: q.Passed ?? null }))] : evaluatedQuestions),
    [evaluatedQuestions, includeAdvanced],
  );

  const questionCount = displayedQuestions.length;
  const evaluatedCount = evaluatedQuestions.length;
  const passedCount = evaluatedQuestions.filter((q) => q.Passed).length;
  const passRate = evaluatedCount ? Math.round((passedCount / evaluatedCount) * 100) : 0;
  const currentTimerValue = quizStarted && !quizFinished ? elapsedTime : lastAttempt?.duration ?? 0;
  const lastAttemptDate = useMemo(
    () => (lastAttempt ? new Date(lastAttempt.completedAt).toLocaleString() : null),
    [lastAttempt],
  );

  const formatTime = (seconds) => {
    const minutes = Math.floor(seconds / 60)
      .toString()
      .padStart(2, '0');
    const remainingSeconds = (seconds % 60).toString().padStart(2, '0');
    return `${minutes}:${remainingSeconds}`;
  };

  return (
    <div className="relative min-h-screen overflow-x-hidden bg-slate-50 text-slate-800">
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute inset-x-0 -top-24 h-72 bg-gradient-to-br from-blue-200 via-sky-100 to-transparent blur-3xl" />
        <div className="absolute inset-y-0 right-0 w-2/3 bg-[radial-gradient(circle_at_top,rgba(37,99,235,0.15),transparent_65%)] blur-2xl" />
      </div>

      <div className="relative z-10 mx-auto flex min-h-screen w-full flex-col gap-12 px-6 pb-16 pt-12 sm:px-10 lg:px-16">
        <header className="rounded-[32px] border border-slate-200 bg-gradient-to-br from-white via-slate-50 to-blue-50/60 p-10 shadow-[0_35px_95px_rgba(15,23,42,0.12)]">
          <div className="space-y-6">
            <span className="inline-flex items-center gap-2 rounded-full bg-blue-100 px-3 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-blue-700 ring-1 ring-inset ring-blue-200">
              Kubernetes mastery lab
            </span>
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl">
              Level up your Kubernetes instincts with immersive scenario drills.
            </h1>
            <p className="max-w-3xl text-base text-slate-600 sm:text-lg">
              Instant access to curated Kubernetes scenarios covering infrastructure builds, security hardening, network policies, audit controls, and operations under pressure. Launch a simulation, apply manifests, and track improvements run after run.
            </p>
            <div className="grid gap-4 sm:grid-cols-3">
              <StatCard
                label="Open challenges"
                value={questionCount || '—'}
                helper={includeAdvanced ? 'Core + advanced scenarios in view.' : 'Core practice set.'}
              />
              <StatCard
                label="Latest score"
                value={lastAttempt ? `${lastAttempt.score}%` : '—'}
                helper={lastAttemptDate ? `Completed ${lastAttemptDate}` : 'Finish a run to collect data.'}
                accent={lastAttempt && lastAttempt.score >= PASS_THRESHOLD ? 'text-blue-600' : ''}
              />
              <StatCard
                label={quizStarted && !quizFinished ? 'Live timer' : 'Last run duration'}
                value={formatTime(currentTimerValue)}
                helper={quizStarted && !quizFinished ? 'Tracking active attempt.' : 'Timer resets on next run.'}
              />
            </div>
            <div className="flex flex-wrap items-center gap-4 pt-4">
              <button
                type="button"
                onClick={startQuiz}
                className={primaryButtonClasses}
                disabled={isLoading}
                aria-label={quizStarted ? 'Restart quiz simulation' : 'Start quiz simulation'}
              >
                {quizStarted ? 'Restart quiz' : 'Start quiz'}
              </button>
              {(quizStarted || quizFinished) && (
                <button type="button" onClick={resetQuiz} className={secondaryButtonClasses} aria-label="Reset quiz session">
                  Reset session
                </button>
              )}
              {isFallback && <span className="text-xs uppercase tracking-[0.3em] text-amber-500">Fallback dataset</span>}
            </div>
          </div>
        </header>

        {quizStarted && !quizFinished && (
          <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-[0_30px_75px_rgba(15,23,42,0.12)]">
            <div className="grid gap-6 lg:grid-cols-[minmax(360px,0.55fr)_minmax(420px,0.45fr)] xl:grid-cols-[minmax(380px,0.5fr)_minmax(460px,0.5fr)]">
              <div className="flex max-h-[760px] flex-col gap-5 overflow-y-auto pr-2">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h2 className="text-2xl font-semibold text-slate-900">Live challenge board</h2>
                    <p className="text-sm text-slate-600">
                      Preview each prompt, map the solution, and iterate in a real cluster or the practice terminal.
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-3">
                    <span className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-sm font-semibold text-slate-600">
                      <svg className="h-4 w-4 text-blue-500" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6l3 3" />
                        <path strokeLinecap="round" strokeLinejoin="round" d="M21 12a9 9 0 1 1-9-9" />
                      </svg>
                      {formatTime(elapsedTime)}
                    </span>
                    <span className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-sm font-semibold text-slate-600">
                      <svg className="h-4 w-4 text-sky-500" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" d="m3 7 9-4 9 4-9 4-9-4v10l9 4 9-4V7" />
                      </svg>
                      {questionCount} challenges
                    </span>
                    <button
                      type="button"
                      onClick={finishQuiz}
                      className={primaryButtonClasses}
                      disabled={isFinishing || isLoading}
                      aria-label="Finish current attempt"
                    >
                      {isFinishing ? 'Submitting…' : 'Finish attempt'}
                    </button>
                  </div>
                </div>

                <div className="mt-6 space-y-4">
                  {error && (
                    <div
                      className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-600"
                      role="alert"
                      aria-live="polite"
                    >
                      {error}
                    </div>
                  )}

                  {isLoading && (
                    <div className="grid gap-3">
                      {Array.from({ length: 4 }).map((_, idx) => (
                        <div key={idx} className="h-24 animate-pulse rounded-2xl bg-slate-100" />
                      ))}
                    </div>
                  )}

                  {!isLoading && questionCount > 0 && (
                    <div className="grid gap-3 md:grid-cols-1" role="list" aria-label="Kubernetes challenges">
                      {displayedQuestions.map((question, index) => (
                        <QuestionCard key={`${question.TestName}-${index}`} question={question} index={index} showResult={false} />
                      ))}
                    </div>
                  )}
                </div>

                <button
                  type="button"
                  onClick={() => setIncludeAdvanced((prev) => !prev)}
                  className="mt-4 inline-flex items-center gap-2 self-start rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-600 transition hover:border-blue-300 hover:text-blue-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400"
                  aria-pressed={includeAdvanced}
                >
                  <span className="inline-flex h-2 w-2 rounded-full bg-blue-500 opacity-80" aria-hidden="true" />
                  {includeAdvanced ? 'Hide advanced scenarios' : 'Show advanced scenarios'}
                </button>
              </div>

              <div className="flex max-h-[760px] justify-center overflow-hidden lg:justify-start">
                <div className="w-full min-w-[360px] lg:min-w-[420px]">
                  <TerminalPanel isVisible={quizStarted && !quizFinished} />
                </div>
              </div>
            </div>
          </section>
        )}

        {quizFinished && (
          <section className="rounded-3xl border border-slate-200 bg-white p-8 shadow-[0_32px_80px_rgba(15,23,42,0.1)]">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 className="text-2xl font-semibold text-slate-900">Results overview</h2>
                <p className="text-sm text-slate-600">Passing threshold is {PASS_THRESHOLD}% — aim for higher on your next run.</p>
              </div>
              <span
                className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-sm font-semibold ${
                  score >= PASS_THRESHOLD
                    ? 'bg-blue-100 text-blue-700 ring-1 ring-inset ring-blue-200'
                    : 'bg-rose-100 text-rose-700 ring-1 ring-inset ring-rose-200'
                }`}
              >
                {score >= PASS_THRESHOLD ? 'Status: Passed' : 'Status: Keep practising'}
              </span>
            </div>

            <div className="mt-6 grid gap-4 sm:grid-cols-3">
              <StatCard
                label="Final score"
                value={`${score}%`}
                helper={`${passedCount} of ${evaluatedCount} challenges cleared.`}
                accent={score >= PASS_THRESHOLD ? 'text-blue-600' : 'text-rose-600'}
              />
              <StatCard label="Time spent" value={formatTime(elapsedTime)} helper="Measured from quiz start." />
              <StatCard label="Pass rate" value={`${passRate || 0}%`} helper="Derived from backend feedback." />
            </div>

            <div className="mt-6 space-y-4">
              {displayedQuestions.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-slate-200 bg-slate-50 px-5 py-10 text-center text-sm text-slate-500">
                  No questions were returned for this attempt.
                </div>
              ) : (
                <div className="grid gap-4 lg:grid-cols-2">
                  {displayedQuestions.map((question, index) => (
                    <QuestionCard key={`${question.TestName}-${index}`} question={question} index={index} showResult />
                  ))}
                </div>
              )}
            </div>

            <div className="mt-6 flex flex-wrap items-center gap-3">
              <button type="button" onClick={startQuiz} className={primaryButtonClasses}>
                Run another simulation
              </button>
              <button type="button" onClick={resetQuiz} className={secondaryButtonClasses}>
                Reset session
              </button>
            </div>
          </section>
        )}

        <footer className="pb-6 text-center text-xs text-slate-400">
          Crafted with care for SREs, platform engineers, and anyone taming clusters in production.
        </footer>
      </div>
    </div>
  );
}

export default App;

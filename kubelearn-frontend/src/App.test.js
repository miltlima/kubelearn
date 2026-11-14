import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

const jsonResponse = (body) => Promise.resolve({ ok: true, json: () => Promise.resolve(body) });

beforeEach(() => {
  // The terminal panel opens a WebSocket as soon as the quiz starts.
  global.WebSocket = jest.fn(() => ({ close: jest.fn(), send: jest.fn() }));
  global.fetch = jest.fn();
  jest.spyOn(console, 'warn').mockImplementation(() => {});
});

test('shows the start button and the question list before an attempt starts', () => {
  render(<App />);

  expect(screen.getByRole('button', { name: /start quiz simulation/i })).toBeInTheDocument();
  expect(screen.getAllByRole('listitem')).toHaveLength(26);
  // Resource names written in backticks are shown as code.
  expect(screen.getByRole('heading', { name: /create a namespace called\s*europe/i })).toBeInTheDocument();
  expect(screen.getByText('europe').tagName).toBe('CODE');
  // Only the catalog is requested; the cluster is not evaluated yet.
  expect(global.fetch).toHaveBeenCalledTimes(1);
  expect(global.fetch).toHaveBeenCalledWith('http://localhost:8083/catalog');
});

test('loads the questions from the backend when the quiz starts', async () => {
  global.fetch.mockImplementation(() =>
    jsonResponse([{ TestName: 'Question 1 - Backend only task', Passed: false, Difficulty: 'Easy' }]),
  );
  render(<App />);

  userEvent.click(screen.getByRole('button', { name: /start quiz simulation/i }));

  expect(await screen.findByRole('heading', { name: /backend only task/i })).toBeInTheDocument();
  expect(global.fetch).toHaveBeenCalledWith('http://localhost:8083/questions');
  expect(screen.getAllByRole('listitem')).toHaveLength(1);
  expect(screen.queryByText(/built-in question list/i)).not.toBeInTheDocument();
});

test('keeps the fallback questions and warns when the backend is unreachable', async () => {
  global.fetch.mockImplementation(() => Promise.reject(new Error('connection refused')));
  render(<App />);

  userEvent.click(screen.getByRole('button', { name: /start quiz simulation/i }));

  expect(await screen.findByRole('alert')).toHaveTextContent(/unable to sync the questions/i);
  expect(screen.getByText(/built-in question list/i)).toBeInTheDocument();
  expect(screen.getAllByRole('listitem')).toHaveLength(26);
});

test('shows the results returned by /finish, not the snapshot from quiz start', async () => {
  const question = { TestName: 'Question 1 - Backend only task', Difficulty: 'Easy' };
  global.fetch.mockImplementation((url) =>
    url.endsWith('/finish')
      ? jsonResponse({ score: 100, results: [{ ...question, Passed: true }] })
      : jsonResponse([{ ...question, Passed: false }]),
  );
  render(<App />);

  userEvent.click(screen.getByRole('button', { name: /start quiz simulation/i }));
  await screen.findByRole('heading', { name: /backend only task/i });
  userEvent.click(screen.getByRole('button', { name: /finish current attempt/i }));

  expect(await screen.findByText(/results overview/i)).toBeInTheDocument();
  await waitFor(() => expect(screen.getByText('1 of 1 questions solved.')).toBeInTheDocument());
  expect(screen.getByText('Status: Passed')).toBeInTheDocument();
  expect(screen.queryByText('Review needed')).not.toBeInTheDocument();
});

describe('check progress', () => {
  const solved = { TestName: 'Question 1 - Solved task', Difficulty: 'Easy', Passed: true };
  const unsolved = {
    TestName: 'Question 2 - Unsolved task',
    Difficulty: 'Medium',
    Passed: false,
    Feedback: 'Replicas is 1, expected 4.',
  };
  const startButton = () => screen.getByRole('button', { name: /start quiz simulation/i });
  const checkButton = () => screen.getByRole('button', { name: /check progress on the challenges/i });

  test('hides pass/fail until the first check', async () => {
    global.fetch.mockImplementation(() => jsonResponse([solved, unsolved]));
    render(<App />);

    userEvent.click(startButton());
    await screen.findByRole('heading', { name: /unsolved task/i });

    expect(screen.queryByText('Passed')).not.toBeInTheDocument();
    expect(screen.queryByText('Not solved yet')).not.toBeInTheDocument();
    expect(screen.queryByText(unsolved.Feedback)).not.toBeInTheDocument();
  });

  test('shows which challenges are solved and why the others are not', async () => {
    global.fetch.mockImplementation(() => jsonResponse([{ ...solved, Passed: false, Feedback: 'Pod not found.' }, unsolved]));
    render(<App />);
    userEvent.click(startButton());
    await screen.findByRole('heading', { name: /unsolved task/i });

    // The learner solves question 1 in the cluster, then checks.
    global.fetch.mockImplementation(() => jsonResponse([solved, unsolved]));
    userEvent.click(checkButton());

    expect(await screen.findByText(/1 of 2 solved/i)).toBeInTheDocument();
    expect(screen.getByText('Passed')).toBeInTheDocument();
    expect(screen.getByText('Not solved yet')).toBeInTheDocument();
    expect(screen.getByText(unsolved.Feedback)).toBeInTheDocument();
    expect(screen.queryByText('Pod not found.')).not.toBeInTheDocument();
    // Checking does not end the attempt.
    expect(screen.queryByText(/results overview/i)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /finish current attempt/i })).toBeInTheDocument();
  });

  test('keeps the challenge list and reports the error when the check fails', async () => {
    global.fetch.mockImplementation(() => jsonResponse([solved, unsolved]));
    render(<App />);
    userEvent.click(startButton());
    await screen.findByRole('heading', { name: /unsolved task/i });

    global.fetch.mockImplementation(() => Promise.resolve({ ok: false, status: 504 }));
    userEvent.click(checkButton());

    expect(await screen.findByRole('alert')).toHaveTextContent(/unable to check your progress/i);
    expect(screen.getByRole('heading', { name: /unsolved task/i })).toBeInTheDocument();
    expect(screen.queryByText(/checked at/i)).not.toBeInTheDocument();
  });

  test('shows the feedback on the results screen too', async () => {
    global.fetch.mockImplementation((url) =>
      url.endsWith('/finish') ? jsonResponse({ score: 50, results: [solved, unsolved] }) : jsonResponse([solved, unsolved]),
    );
    render(<App />);
    userEvent.click(startButton());
    await screen.findByRole('heading', { name: /unsolved task/i });
    userEvent.click(screen.getByRole('button', { name: /finish current attempt/i }));

    expect(await screen.findByText(/results overview/i)).toBeInTheDocument();
    expect(await screen.findByText(unsolved.Feedback)).toBeInTheDocument();
    expect(screen.getByText('Review needed')).toBeInTheDocument();
  });
});

describe('tracks', () => {
  const catalog = [
    { TestName: 'Question 1 - Core task', Difficulty: 'Easy', Track: 'Core' },
    { TestName: 'Question 27 - Security task', Difficulty: 'Medium', Track: 'CKS' },
    { TestName: 'Question 38 - Network task one', Difficulty: 'Easy', Track: 'CKNE' },
    { TestName: 'Question 39 - Network task two', Difficulty: 'Hard', Track: 'CKNE' },
  ];
  const backend = (url) => {
    if (url.endsWith('/catalog')) return jsonResponse(catalog);
    const track = new URL(url).searchParams.get('track');
    const results = catalog
      .filter((q) => !track || q.Track.toLowerCase() === track)
      .map((q) => ({ ...q, Passed: true }));
    return url.includes('/finish') ? jsonResponse({ score: 100, results }) : jsonResponse(results);
  };

  test('lists the questions from the backend catalog with a count per track', async () => {
    global.fetch.mockImplementation(backend);
    render(<App />);

    expect(await screen.findByRole('heading', { name: /security task/i })).toBeInTheDocument();
    expect(screen.getAllByRole('listitem')).toHaveLength(4);
    expect(screen.getByRole('radio', { name: 'All track, 4 questions' })).toBeChecked();
    expect(screen.getByRole('radio', { name: 'CKS track, 1 questions' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'CKNE track, 2 questions' })).toBeInTheDocument();
  });

  test('choosing a track limits the list and the attempt to it', async () => {
    global.fetch.mockImplementation(backend);
    render(<App />);
    await screen.findByRole('heading', { name: /security task/i });

    userEvent.click(screen.getByRole('radio', { name: 'CKNE track, 2 questions' }));

    expect(screen.getAllByRole('listitem')).toHaveLength(2);
    expect(screen.queryByRole('heading', { name: /security task/i })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /start quiz simulation/i })).toHaveTextContent('Start attempt with 2 questions');

    userEvent.click(screen.getByRole('button', { name: /start quiz simulation/i }));
    // The same headings exist on the start screen, so wait for the attempt view itself.
    await screen.findByRole('button', { name: /finish current attempt/i });
    await waitFor(() => expect(screen.getAllByRole('listitem')).toHaveLength(2));
    expect(global.fetch).toHaveBeenCalledWith('http://localhost:8083/questions?track=ckne');

    userEvent.click(screen.getByRole('button', { name: /finish current attempt/i }));
    expect(await screen.findByText('2 of 2 questions solved.')).toBeInTheDocument();
    expect(global.fetch).toHaveBeenCalledWith('http://localhost:8083/finish?track=ckne');
  });

  test('explains that only the built-in list is available without the backend', async () => {
    global.fetch.mockImplementation(() => Promise.reject(new Error('connection refused')));
    render(<App />);

    expect(await screen.findByText(/backend is not answering/i)).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: 'Core track, 26 questions' })).toBeInTheDocument();
    userEvent.click(screen.getByRole('radio', { name: 'CKS track, 0 questions' }));
    expect(screen.getByRole('button', { name: /start quiz simulation/i })).toBeDisabled();
  });
});

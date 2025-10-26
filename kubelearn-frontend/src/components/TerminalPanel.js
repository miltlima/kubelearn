import React, { useEffect, useRef, useState, useCallback } from 'react';

const STATUS = {
  CONNECTING: 'connecting',
  OPEN: 'open',
  CLOSED: 'closed',
  ERROR: 'error',
};

const TerminalPanel = ({ isVisible }) => {
  const [socket, setSocket] = useState(null);
  const [status, setStatus] = useState(STATUS.CONNECTING);
  const [lines, setLines] = useState([]);
  const [command, setCommand] = useState('');
  const [history, setHistory] = useState([]);
  const [, setHistoryIndex] = useState(-1);
  const [reconnectToken, setReconnectToken] = useState(0);
  const outputRef = useRef(null);

  useEffect(() => {
    if (!isVisible) {
      setSocket((prev) => {
        if (prev) {
          prev.close();
        }
        return null;
      });
      setStatus(STATUS.CLOSED);
      return;
    }

    let ws;
    try {
      ws = new WebSocket('ws://localhost:8083/terminal');
    } catch (connectionError) {
      console.warn('[KubeLearn] Unable to create terminal WebSocket:', connectionError);
      setStatus(STATUS.ERROR);
      setLines((prev) => [...prev, 'Could not open terminal connection. Check if the backend service is running on port 8083.']);
      return;
    }

    setSocket(ws);
    setStatus(STATUS.CONNECTING);

    ws.onopen = () => setStatus(STATUS.OPEN);
    ws.onclose = () => setStatus(STATUS.CLOSED);
    ws.onerror = (event) => {
      const detail = event?.message || '';
      console.warn(`[KubeLearn] Terminal connection error${detail ? `: ${detail}` : ''}`, event);
      setStatus(STATUS.ERROR);
      setLines((prev) => [...prev, 'Connection error. Verify the backend service is running on port 8083.']);
    };
    ws.onmessage = (event) => {
      setLines((prev) => {
        const incoming = event.data.trim();
        if (!incoming) {
          return prev;
        }
        const lastLine = prev[prev.length - 1];
        if (lastLine && lastLine.trim() === incoming) {
          return prev;
        }
        return [...prev, incoming];
      });
    };

    return () => {
      ws.close();
      setSocket(null);
    };
  }, [isVisible, reconnectToken]);

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [lines]);

  const clearTerminal = useCallback(() => {
    setLines([]);
    setHistoryIndex(-1);
  }, []);

  const sendCommand = useCallback(
    (value) => {
      if (!value.trim() || !socket || status !== STATUS.OPEN) {
        return;
      }
      if (value.trim().toLowerCase() === 'clear') {
        clearTerminal();
        setCommand('');
        setHistory((prev) => ['clear', ...prev].slice(0, 50));
        return;
      }
      const nextLines = [...lines, `> ${value}`];
      setLines(nextLines);
      socket.send(`${value}\n`);
      setHistory((prev) => [value, ...prev].slice(0, 50));
      setHistoryIndex(-1);
      setCommand('');
    },
    [clearTerminal, lines, socket, status],
  );

  const handleSubmit = (event) => {
    event.preventDefault();
    sendCommand(command);
  };

  const handleKeyDown = (event) => {
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      setHistoryIndex((prev) => {
        const nextIndex = Math.min(prev + 1, history.length - 1);
        const nextCommand = history[nextIndex] ?? command;
        setCommand(nextCommand);
        return nextIndex;
      });
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      setHistoryIndex((prev) => {
        if (prev <= 0) {
          setCommand('');
          return -1;
        }
        const nextIndex = prev - 1;
        setCommand(history[nextIndex] ?? '');
        return nextIndex;
      });
    }
  };

  const handleReconnect = () => {
    if (socket) {
      socket.close();
    }
    setLines((prev) => [...prev, '--- reconnecting terminal session ---']);
    setReconnectToken((prev) => prev + 1);
  };

  const statusText = {
    [STATUS.CONNECTING]: 'Connecting',
    [STATUS.OPEN]: 'Connected',
    [STATUS.CLOSED]: 'Closed',
    [STATUS.ERROR]: 'Error',
  }[status];

  const statusBadgeClasses = {
    [STATUS.CONNECTING]: 'bg-amber-100 text-amber-700 ring-1 ring-inset ring-amber-200',
    [STATUS.OPEN]: 'bg-blue-100 text-blue-700 ring-1 ring-inset ring-blue-200',
    [STATUS.CLOSED]: 'bg-slate-100 text-slate-600 ring-1 ring-inset ring-slate-200',
    [STATUS.ERROR]: 'bg-rose-100 text-rose-700 ring-1 ring-inset ring-rose-200',
  }[status];

  const canReconnect = status === STATUS.ERROR || status === STATUS.CLOSED;

  return (
    <div className="flex h-[720px] w-full flex-col overflow-hidden rounded-3xl border border-slate-200 bg-gradient-to-br from-white via-slate-50 to-blue-50 p-6 font-mono shadow-[0_30px_80px_rgba(15,23,42,0.12)] lg:h-[760px]">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-white px-3 py-1">
            <div className="flex gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-rose-400" />
              <span className="h-2.5 w-2.5 rounded-full bg-amber-400" />
              <span className="h-2.5 w-2.5 rounded-full bg-blue-400" />
            </div>
            <span className="text-[0.65rem] font-semibold uppercase tracking-[0.32em] text-slate-500">Terminal</span>
          </div>
          <span className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold ${statusBadgeClasses}`}>
            {statusText}
          </span>
          {canReconnect && (
            <button
              type="button"
              onClick={handleReconnect}
              className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-600 transition hover:border-blue-300 hover:text-blue-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400"
            >
              Reconnect
            </button>
          )}
        </div>
        <button
          type="button"
          onClick={() => {
            clearTerminal();
            setCommand('');
          }}
          className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-semibold text-slate-600 transition hover:border-blue-300 hover:text-blue-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400"
        >
          Clear
        </button>
      </header>

      <div
        ref={outputRef}
        className="mt-5 flex-1 overflow-y-auto rounded-2xl border border-slate-200 bg-slate-50 p-4 font-mono text-sm text-slate-700 shadow-inner shadow-slate-200/70"
      >
        {lines.length === 0 ? (
          <p className="text-slate-500">> Awaiting command… try `kubectl get pods`</p>
        ) : (
          lines.map((line, index) => (
            <pre key={`${line}-${index}`} className="whitespace-pre-wrap text-slate-700">
              {line}
            </pre>
          ))
        )}
      </div>

      <form onSubmit={handleSubmit} className="mt-4">
        <div className="flex items-center gap-3 rounded-full border border-slate-200 bg-white px-4 shadow-sm">
          <span className="text-sm font-semibold text-blue-600">kubelearner$</span>
          <input
            type="text"
            value={command}
            onChange={(event) => setCommand(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a command and press Enter"
            className="flex-1 bg-transparent py-3 text-sm text-slate-700 placeholder:text-slate-400 focus:outline-none"
            disabled={status !== STATUS.OPEN}
            aria-label="Terminal command input"
          />
        </div>
      </form>
    </div>
  );
};

export default TerminalPanel;

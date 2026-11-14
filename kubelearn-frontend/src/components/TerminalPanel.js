import React, { useEffect, useRef, useState, useCallback } from 'react';
import { translator } from '../i18n';

const STATUS = {
  CONNECTING: 'connecting',
  OPEN: 'open',
  CLOSED: 'closed',
  ERROR: 'error',
};

// Used when no translator is passed in: show the key's English text.
const english = translator('en');

const TerminalPanel = ({ isVisible, t = english }) => {
  // The connection effect must not restart when the language changes, so its
  // messages read the current translator through a ref.
  const translate = useRef(t);
  translate.current = t;

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
      setLines((prev) => [...prev, translate.current('terminal.cannotOpen')]);
      return;
    }

    setSocket(ws);
    setStatus(STATUS.CONNECTING);

    // Once this effect is cleaned up the socket is stale: its late events
    // (for example the error from closing it while still connecting) must not
    // touch the panel, which by then belongs to a newer connection.
    let stale = false;

    ws.onopen = () => {
      if (!stale) setStatus(STATUS.OPEN);
    };
    ws.onclose = () => {
      if (!stale) setStatus(STATUS.CLOSED);
    };
    ws.onerror = (event) => {
      if (stale) return;
      const detail = event?.message || '';
      console.warn(`[KubeLearn] Terminal connection error${detail ? `: ${detail}` : ''}`, event);
      setStatus(STATUS.ERROR);
      setLines((prev) => [...prev, translate.current('terminal.connectionError')]);
    };
    ws.onmessage = (event) => {
      if (stale) return;
      setLines((prev) => {
        const incoming = event.data.trim();
        if (!incoming) {
          return prev;
        }
        return [...prev, incoming];
      });
    };

    return () => {
      stale = true;
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
    setLines((prev) => [...prev, t('terminal.reconnecting')]);
    setReconnectToken((prev) => prev + 1);
  };

  const statusText = {
    [STATUS.CONNECTING]: t('terminal.connecting'),
    [STATUS.OPEN]: t('terminal.open'),
    [STATUS.CLOSED]: t('terminal.closed'),
    [STATUS.ERROR]: t('terminal.error'),
  }[status];

  const statusDot = {
    [STATUS.CONNECTING]: 'bg-warn',
    [STATUS.OPEN]: 'bg-pass',
    [STATUS.CLOSED]: 'bg-muted',
    [STATUS.ERROR]: 'bg-fail',
  }[status];

  const canReconnect = status === STATUS.ERROR || status === STATUS.CLOSED;
  const chromeButton =
    'rounded px-2 py-1 text-sm text-muted transition-colors hover:text-fg focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-kube-soft';

  return (
    <div className="flex h-full min-h-0 flex-col bg-term">
      <header className="flex items-center gap-4 border-b border-line px-4 py-2">
        <h2 className="text-sm font-semibold text-fg">{t('terminal.title')}</h2>
        <p className="flex items-center gap-2 text-sm text-muted">
          <span className={`h-2 w-2 rounded-full ${statusDot}`} aria-hidden="true" />
          {statusText}
        </p>
        {canReconnect && (
          <button type="button" onClick={handleReconnect} className={chromeButton}>
            {t('terminal.reconnect')}
          </button>
        )}
        <button
          type="button"
          onClick={() => {
            clearTerminal();
            setCommand('');
          }}
          className={`ml-auto ${chromeButton}`}
        >
          {t('terminal.clear')}
        </button>
      </header>

      <div
        ref={outputRef}
        role="log"
        aria-label={t('terminal.output')}
        className="min-h-0 flex-1 overflow-auto px-4 py-3 font-mono text-[13px] leading-relaxed text-fg"
      >
        {lines.length === 0 ? (
          <p className="text-muted">{t('terminal.hint')}</p>
        ) : (
          lines.map((line, index) => (
            <pre
              key={`${line}-${index}`}
              className={`w-max whitespace-pre font-mono ${line.startsWith('> ') ? 'mt-3 text-kube-soft first:mt-0' : ''}`}
            >
              {line}
            </pre>
          ))
        )}
      </div>

      <form onSubmit={handleSubmit} className="flex items-center gap-3 border-t border-line px-4 py-3 font-mono text-[13px]">
        <span className="font-medium text-kube-soft" aria-hidden="true">
          $
        </span>
        <input
          type="text"
          value={command}
          onChange={(event) => setCommand(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={status === STATUS.OPEN ? t('terminal.placeholder') : t('terminal.disconnected')}
          className="min-w-0 flex-1 bg-transparent text-fg placeholder:text-muted/70 focus:outline-none"
          disabled={status !== STATUS.OPEN}
          aria-label={t('terminal.input')}
          autoComplete="off"
          autoCapitalize="off"
          spellCheck={false}
        />
      </form>
    </div>
  );
};

export default TerminalPanel;

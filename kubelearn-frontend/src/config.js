// Where the backend lives.
//
// The released binary serves this interface itself, so a production build
// talks to the address it was loaded from. During development the interface
// runs on its own dev server and the backend on port 8083.
// REACT_APP_API_URL overrides both.
const sameOrigin = process.env.NODE_ENV === 'production';

export const API = process.env.REACT_APP_API_URL ?? (sameOrigin ? '' : 'http://localhost:8083');

// WebSocket address of the terminal: the API address with ws:// (or wss://).
export const terminalURL = () => {
  const base = API || window.location.origin;
  return `${base.replace(/^http/, 'ws')}/terminal`;
};

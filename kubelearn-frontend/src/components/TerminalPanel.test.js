import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TerminalPanel from './TerminalPanel';

let socket;

beforeEach(() => {
  socket = null;
  global.WebSocket = jest.fn(() => {
    socket = { close: jest.fn(), send: jest.fn() };
    return socket;
  });
  jest.spyOn(console, 'warn').mockImplementation(() => {});
});

const openTerminal = () => {
  render(<TerminalPanel isVisible />);
  act(() => socket.onopen());
};

const runCommand = (command, output) => {
  userEvent.type(screen.getByLabelText(/terminal command input/i), `${command}{enter}`);
  act(() => socket.onmessage({ data: output }));
};

test('sends the typed command to the backend', () => {
  openTerminal();

  runCommand('kubectl get pods', 'pod/nginx is Running');

  expect(socket.send).toHaveBeenCalledWith('kubectl get pods\n');
  expect(screen.getByText('> kubectl get pods')).toBeInTheDocument();
  expect(screen.getByText('pod/nginx is Running')).toBeInTheDocument();
});

test('shows the output again when the same command returns the same result', () => {
  openTerminal();

  runCommand('kubectl get pods', 'pod/nginx is Running');
  runCommand('kubectl get pods', 'pod/nginx is Running');

  expect(screen.getAllByText('> kubectl get pods')).toHaveLength(2);
  expect(screen.getAllByText('pod/nginx is Running')).toHaveLength(2);
});

test('shows identical consecutive outputs from different commands', () => {
  openTerminal();

  runCommand('kubectl get pods -n a', 'No resources found');
  act(() => socket.onmessage({ data: 'No resources found' }));

  expect(screen.getAllByText('No resources found')).toHaveLength(2);
});

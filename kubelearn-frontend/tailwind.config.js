/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#0B1220', // page background
        panel: '#111A2E', // bars and raised surfaces
        term: '#070C17', // terminal background
        line: '#22304D', // borders and dividers
        fg: '#E6ECF7', // primary text
        muted: '#8FA0BF', // secondary text
        kube: '#326CE5', // Kubernetes blue: primary actions
        'kube-soft': '#7FA8FF', // links, focus, prompt
        pass: '#3FB68B',
        fail: '#F0787C',
        warn: '#E2B25B',
      },
      fontFamily: {
        sans: ['"IBM Plex Sans"', 'system-ui', 'sans-serif'],
        mono: ['"IBM Plex Mono"', 'ui-monospace', 'SFMono-Regular', 'monospace'],
      },
    },
  },
  plugins: [],
};

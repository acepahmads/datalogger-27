/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        operator: {
          bg: '#050B16',
          panel: '#0B1422',
          panelSecondary: '#101B2B',
          border: '#20354B',
          primary: '#168BFF',
          cyan: '#00CFFF',
          normal: '#00D6A0',
          warning: '#FFB020',
          critical: '#FF4559',
          text: '#F1F5F9',
          muted: '#94A3B8',
        },
        dark: {
          bg: '#0B0F19',
          surface: '#111827',
          card: '#131B2E',
          cardHover: '#17223B',
          border: '#1E293B',
          borderSoft: '#1A2338',
          borderHighlight: '#2A3756',
        },
        brand: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#0B0F19',
        },
        skytheme: {
          bg: '#F4FAFF',
          tint: '#EAF5FF',
          accent: '#D9EEFF',
          surface: '#FFFFFF',
          primary: '#1685E5',
          text: '#15324F',
          secondary: '#58718D',
          border: '#B8D4EA',
          borderStrong: '#94AEC5',
        },
        status: {
          done: '#10b981',
          working: '#3b82f6',
          testing: '#a855f7',
          blocked: '#ef4444',
          waiting: '#f59e0b',
          pending: '#64748b',
        }
      },
      fontFamily: {
        sans: ['Inter', 'Manrope', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'Courier New', 'monospace'],
      },
      fontSize: {
        '2xs': '0.6875rem', // 11px
        '3xs': '0.625rem',  // 10px
        '4xs': '0.5625rem', // 9px
        '5xs': '0.5rem',    // 8px
      },
      borderRadius: {
        'xl': '0.75rem',
        '2xl': '1rem',
      },
      boxShadow: {
        'soft': '0 2px 10px rgba(0, 0, 0, 0.25)',
        'glow-blue': '0 0 15px rgba(37, 99, 235, 0.2)',
        'glow-emerald': '0 0 15px rgba(16, 185, 129, 0.2)',
      }
    },
  },
  plugins: [],
}

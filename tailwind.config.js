/** @type {import('tailwindcss').Config} */
import typography from '@tailwindcss/typography';

export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: { indigo: '#4f46e5', purple: '#9333ea', fuchsia: '#c026d3' },
      },
      backgroundImage: {
        'brand-gradient': 'linear-gradient(135deg,#4f46e5 0%,#9333ea 55%,#c026d3 100%)',
        'brand-radial': 'radial-gradient(60% 60% at 50% 0%,rgba(147,51,234,.35),transparent 70%)',
        // Split-complementary accents: teal (secondary) and amber (rare pop)
        'accent-teal': 'linear-gradient(135deg,#14b8a6 0%,#06b6d4 100%)',
        'accent-amber': 'linear-gradient(135deg,#f59e0b 0%,#f97316 100%)',
      },
      boxShadow: { 'brand-glow': '0 10px 40px -10px rgba(99,51,234,.55)' },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      keyframes: {
        'gradient-pan': { '0%,100%': { backgroundPosition: '0% 50%' }, '50%': { backgroundPosition: '100% 50%' } },
        'float': { '0%,100%': { transform: 'translateY(0)' }, '50%': { transform: 'translateY(-8px)' } },
        'aurora': { '0%': { transform: 'translate(-10%,-10%) rotate(0deg)' }, '100%': { transform: 'translate(10%,10%) rotate(8deg)' } },
        'fade-in': { '0%': { opacity: '0', transform: 'translateY(8px)' }, '100%': { opacity: '1', transform: 'translateY(0)' } },
      },
      animation: {
        'gradient-pan': 'gradient-pan 8s ease infinite',
        'float': 'float 6s ease-in-out infinite',
        'aurora': 'aurora 18s ease-in-out infinite alternate',
        'fade-in': 'fade-in 200ms ease-out',
      },
      typography: {
        DEFAULT: {
          css: {
            h3: {
              marginTop: '0.6em',
            },
            "p a": {
              textDecoration: 'underline',
            },
          },
        },
      },
    },
  },
  plugins: [typography],
}

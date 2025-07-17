/** @type {import('tailwindcss').Config} */
import typography from '@tailwindcss/typography';

export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
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


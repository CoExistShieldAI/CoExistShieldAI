/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,ts}",
  ],
  theme: {
    extend: {
      colors: {
        tier: {
          safe: '#10B981',
          notice: '#38BDF8',
          warning: '#F59E0B',
          danger: '#EF4444',
        }
      }
    },
  },
  plugins: [],
}

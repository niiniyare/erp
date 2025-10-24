module.exports = {
  content: [
    "./web/**/*.templ",
    "./web/components/**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        primary: {...},    // Brand colors
        secondary: {...},
      },
      spacing: {
        '18': '4.5rem',   // Custom spacing
      },
    },
  },
  plugins: [
    require('flowbite/plugin'),
    require('@tailwindcss/forms'),
  ],
}

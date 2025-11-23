/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                gray: {
                    900: 'rgb(var(--color-gray-900) / <alpha-value>)',
                    800: 'rgb(var(--color-gray-800) / <alpha-value>)',
                    750: 'rgb(var(--color-gray-750) / <alpha-value>)',
                    700: 'rgb(var(--color-gray-700) / <alpha-value>)',
                    600: 'rgb(var(--color-gray-600) / <alpha-value>)',
                    500: 'rgb(var(--color-gray-500) / <alpha-value>)',
                    400: 'rgb(var(--color-gray-400) / <alpha-value>)',
                    300: 'rgb(var(--color-gray-300) / <alpha-value>)',
                    200: 'rgb(var(--color-gray-200) / <alpha-value>)',
                    100: 'rgb(var(--color-gray-100) / <alpha-value>)',
                },
                blue: {
                    700: 'rgb(var(--color-blue-700) / <alpha-value>)',
                    600: 'rgb(var(--color-blue-600) / <alpha-value>)',
                    500: 'rgb(var(--color-blue-500) / <alpha-value>)',
                    400: 'rgb(var(--color-blue-400) / <alpha-value>)',
                },
                primary: '#0f172a', // Keep for backward compatibility if used
                secondary: '#1e293b',
                accent: '#3b82f6',
            },
            borderRadius: {
                'none': '0',
                'sm': 'var(--radius-sm)',
                DEFAULT: 'var(--radius-md)',
                'md': 'var(--radius-md)',
                'lg': 'var(--radius-lg)',
                'full': '9999px',
            },
        },
    },
    plugins: [],
}

/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			fontFamily: {
				sans: [
					'Inter var, sans-serif',
					{
						fontFeatureSettings: '"cv11", "ss01"'
					}
				]
			}
		}
	},
	plugins: [require('@tailwindcss/forms')]
};

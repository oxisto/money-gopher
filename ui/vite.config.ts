import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import houdini from 'houdini/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		houdini(),
		tailwindcss(),
		sveltekit({
			alias: {
				$houdini: './.houdini',
				'$houdini/*': './.houdini/*'
			},
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// The UI is a pure SPA: no SSR, everything falls back to index.html,
			// so it can later be embedded into the moneyd binary via go:embed.
			adapter: adapter({ fallback: 'index.html' })
		})
	],
	server: {
		// The dev server proxies GraphQL to a locally running moneyd.
		proxy: {
			'/graphql': 'http://localhost:8080'
		}
	}
});

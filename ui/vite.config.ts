import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/mgo.portfolio.v1.PortfolioService': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false
			},
			'/mgo.portfolio.v1.SecuritiesService': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false
			}
		}
	}
});

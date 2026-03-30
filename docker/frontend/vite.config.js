import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
	plugins: [react()],
	base: '/app/',   // ✅ comma added here
	test: {
		globals: true,
		environment: 'jsdom',
		setupFiles: './src/setupTests.js',
		coverage: {
			reporter: ['text', 'lcov'],
		},
	},
})

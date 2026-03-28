import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
	plugins: [react()],
	base: '/app/'   // <-- this ensures JS/CSS assets resolve correctly
})

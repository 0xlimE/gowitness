import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/', // Ensure assets use absolute paths from root
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});

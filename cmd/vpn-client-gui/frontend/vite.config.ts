import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    strictPort: true,
  },
});

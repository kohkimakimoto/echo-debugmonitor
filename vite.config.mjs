import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import templateExtract from './vite-plugin-template-extract.js';

export default defineConfig({
  publicDir: false,
  plugins: [
    templateExtract(),
    tailwindcss(),
  ],
  build: {
    manifest: "manifest.json",
    outDir: "public/build",
    rollupOptions: {
      input: ['assets/app.js'],
    },
  },
});

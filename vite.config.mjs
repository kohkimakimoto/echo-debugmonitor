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
    outDir: "resources/public/build",
    rollupOptions: {
      input: ['resources/assets/app.js'],
      onwarn(warning, warn) {
        // Suppress eval warnings from htmx
        if (warning.code === 'EVAL') {
          return;
        }
        // Pass other warnings to the default handler
        warn(warning);
      }
    },
  },
});

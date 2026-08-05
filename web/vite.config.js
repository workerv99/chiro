import { resolve } from "path";
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()],
  resolve: {
    alias: {
      "$lib": resolve("./src/lib"),
      "$components": resolve("./src/lib/components"),
      "$utils": resolve("./src/lib/utils")
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8081'
    }
  }
});

import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      fallback: '200.html',
      precompress: true
    }),
    version: {
      // Use package version or build timestamp for version tracking
      name: process.env.npm_package_version || Date.now().toString(),
      // Check for new version every 60 seconds
      pollInterval: 60000
    }
  },
  compilerOptions: {
    experimental: {
      async: true
    }
  }
};

export default config;

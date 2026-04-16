// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from "@tailwindcss/vite";


import svelte from "@astrojs/svelte";

import vue from "@astrojs/vue";

import react from "@astrojs/react";

import mdx from "@astrojs/mdx";

// https://astro.build/config
export default defineConfig({
  vite: {
    plugins: [tailwindcss()],
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:1323',
          changeOrigin: true,
        },
      },
    },
  },

  integrations: [svelte(), vue(), react(), mdx()],
});
import { defineConfig } from 'astro/config';
import node from "@astrojs/node";
import preact from "@astrojs/preact";

import svelte from "@astrojs/svelte";

// https://astro.build/config
export default defineConfig({
  output: "server",
  adapter: node({
    mode: "standalone"
  }),
  integrations: [preact(), svelte()]
});
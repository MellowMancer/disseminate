import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: true,           // Listen on all network interfaces
    watch: {
      usePolling: true,   // Use polling for file watching inside Docker
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});

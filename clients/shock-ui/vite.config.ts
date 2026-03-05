import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: "/ui/",
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/node": "http://localhost:7445",
      "/preauth": "http://localhost:7445",
      "/location": "http://localhost:7445",
      "/locker": "http://localhost:7445",
      "/locked": "http://localhost:7445",
      "/trace": "http://localhost:7445",
      "/types": "http://localhost:7445",
    },
  },
});

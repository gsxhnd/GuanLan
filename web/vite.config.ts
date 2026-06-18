import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

const workspaceRoot = path.resolve(__dirname, "..")

// https://vite.dev/config/
export default defineConfig({
  cacheDir: path.resolve(__dirname, "node_modules/.vite"),
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  optimizeDeps: {
    include: [
      "react",
      "react-dom",
      "react-router",
      "zustand",
      "zustand/middleware",
      "i18next",
      "react-i18next",
    ],
  },
  server: {
    port: 1420,
    strictPort: true,
    fs: {
      allow: [workspaceRoot],
    },
  },
})

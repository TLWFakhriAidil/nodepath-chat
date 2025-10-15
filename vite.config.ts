import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [react()],
  server: {
    host: "::",
    port: 8080,
    allowedHosts: ["nodepath-chat-production.up.railway.app"],
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  },
  preview: {
    host: "0.0.0.0",
    port: 4173,
    allowedHosts: ["nodepath-chat-production.up.railway.app"]
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));

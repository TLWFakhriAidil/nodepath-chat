import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 8080,
    allowedHosts: ["nodepath-chat-production.up.railway.app"],
    proxy: {
      '/mysql-api.php': {
        target: 'https://nodepath-chat-production.up.railway.app',
        changeOrigin: true,
        secure: true
      },
      '/api': {
        target: 'https://nodepath-chat-production.up.railway.app',
        changeOrigin: true,
        secure: true
      }
    }
  },
  preview: {
    host: "0.0.0.0",
    port: 4173,
    allowedHosts: ["nodepath-chat-production.up.railway.app"]
  },
  plugins: [
    react(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));

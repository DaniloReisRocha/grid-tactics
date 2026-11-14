import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    // Expose the development server so local tunnels (such as ngrok) can reach it.
    host: "0.0.0.0",
    // Vite 7 rejects unknown Host headers by default. Keep the exception scoped
    // to ngrok domains instead of disabling host validation entirely.
    allowedHosts: [".ngrok-free.app", ".ngrok.app"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});

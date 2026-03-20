import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: ["rentcar-app-bdcqacg7g7bag9fc.germanywestcentral-01.azurewebsites.net"],
  },
});

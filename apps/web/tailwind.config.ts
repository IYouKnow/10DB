import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#08111f",
        mist: "#edf4ff",
        glow: "#8dd2c8",
        ember: "#ff9f68",
        panel: "#0f1b2d"
      },
      boxShadow: {
        panel: "0 22px 60px rgba(4, 12, 24, 0.22)"
      }
    }
  },
  plugins: []
} satisfies Config;

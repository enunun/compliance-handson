import { serve } from "@hono/node-server";
import { createApp } from "./app.tsx";

const port = Number(process.env.PORT ?? 3000);
const app = createApp({
  apiUrl: process.env.API_URL ?? "http://localhost:8080",
});

serve({ fetch: app.fetch, port }, () => {
  console.log(`web listening on :${port}`);
});

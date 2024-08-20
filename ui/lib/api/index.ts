import createClient, { Middleware } from "openapi-fetch";
import type { paths } from "./v1";
import { auth } from "@/lib/auth";
export * from "./types";

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    const session = await auth();
    request.headers.set("Authorization", `Bearer ${session?.accessToken}`);
    return request;
  },
};

const client = createClient<paths>({ baseUrl: "http://localhost:8080" });
client.use(authMiddleware);

export default client;

import { decode } from "next-auth/jwt";
import { cookies } from "next/headers";
import createClient, { Middleware } from "openapi-fetch";
import type { paths } from "./v1";
export * from "./types";

const authMiddleware: Middleware = {
  async onRequest({ request }) {
    // Build the cookie name
    const cookieName =
      process.env.NODE_ENV === "production"
        ? "__Secure-authjs.session-token"
        : "authjs.session-token";
    // Retrieve our frontend auth cookie that contains the encrypted frontend
    // token
    const cookie = cookies().get(cookieName);

    // Decode/decrypt the token to access the backend API token
    const token = await decode({
      token: cookie?.value,
      secret: process.env.AUTH_SECRET ?? "",
      salt: cookieName,
    });

    // Set the backend API token
    request.headers.set("Authorization", `Bearer ${token?.backendAPIToken}`);
    return request;
  },
};

const client = createClient<paths>({ baseUrl: "http://localhost:8080" });
client.use(authMiddleware);

export default client;

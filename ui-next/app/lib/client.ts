import { auth } from "@/lib/auth";
import { PortfolioService, SecuritiesService } from "@/lib/gen/mgo_connect";
import type { Interceptor, PromiseClient } from "@connectrpc/connect";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

const authorizer: Interceptor = (next) => async (req) => {
  const session = await auth();
  console.log(session);
  req.header.set("Authorization", `Bearer ${session?.accessToken}`);
  return await next(req);
};

export function portfolioClient(
  fetch = window.fetch
): PromiseClient<typeof PortfolioService> {
  return createPromiseClient(
    PortfolioService,
    createConnectTransport({
      baseUrl: "http://localhost:8080",
      useHttpGet: true,
      fetch: fetch,
      interceptors: [authorizer],
    })
  );
}

export function securitiesClient(
  fetch = window.fetch
): PromiseClient<typeof SecuritiesService> {
  return createPromiseClient(
    SecuritiesService,
    createConnectTransport({
      baseUrl: "http://localhost:8080",
      useHttpGet: true,
      fetch: fetch,
      interceptors: [authorizer],
    })
  );
}

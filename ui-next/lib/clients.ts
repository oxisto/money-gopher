import { auth } from "@/lib/auth";
import { PortfolioService, SecuritiesService } from "@/lib/gen/mgo_connect";
import type { Interceptor } from "@connectrpc/connect";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

const authorizer: Interceptor = (next) => async (req) => {
  const session = await auth();
  req.header.set("Authorization", `Bearer ${session?.accessToken}`);
  return await next(req);
};

export const portfolioClient = createPromiseClient(
  PortfolioService,
  createConnectTransport({
    baseUrl: "http://localhost:8080",
    useHttpGet: true,
    fetch: fetch,
    interceptors: [authorizer],
  })
);

export const securitiesClient = createPromiseClient(
  SecuritiesService,
  createConnectTransport({
    baseUrl: "http://localhost:8080",
    useHttpGet: true,
    fetch: fetch,
    interceptors: [authorizer],
  })
);

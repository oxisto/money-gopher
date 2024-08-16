import NextAuth from "next-auth";

export const { handlers, auth } = NextAuth({
  cookies: {
    pkceCodeVerifier: {
      name: "next-auth.pkce.code_verifier",
      options: {
        httpOnly: true,
        sameSite: "none",
        path: "/",
        secure: true,
      },
    },
  },
  providers: [
    {
      id: "money-gopher",
      name: "My Provider",
      type: "oauth",
      wellKnown: "http://localhost:8000/.well-known/openid-configuration",
      token: {
        url: "http://localhost:8000/token",
      },
      issuer: "http://localhost:8000",
      clientId: process.env.AUTH_CLIENT_ID,
      client: {
        token_endpoint_auth_method: "none",
      },
      userinfo: {
        request: () => {},
      },
      checks: ["pkce"],
    },
  ],
  callbacks: {
    authorized: async ({ auth }) => {
      return !!auth;
    },
    jwt: async ({ token, user, account }) => {
      console.log(`account: ${account}`)
      if (account && account.access_token) {
        // set access_token to the token payload
        token.accessToken = account.access_token;
      }

      return token;
    },
    session: async ({ session, token, user }) => {
      // If we want to make the accessToken available in components, then we have to explicitly forward it here.
      return { ...session, accessToken: token.accessToken };
    },
  },
});

declare module "next-auth" {
  interface Session {
    accessToken?: string
  }
}

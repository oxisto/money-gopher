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
      if (account && account.access_token) {
        // We want to make the backend API token available to our server
        // components, so we include them in the "token" that is used to
        // authenticate backend/frontend.
        //
        // Note: While this means that the API token ends up in this token, the
        // client will not be able to read it since it is encrypted on the
        // client-side.
        token.backendAPIToken = account.access_token;

        // Sync the expiry of our backend token with the frontend token
        token.exp = account.expires_at;
      }

      return token;
    },
    session: async ({ session, token, user }) => {
      return {
        ...session,
        user: {
          // Until we have a proper user-info endpoint, just set a fake name
          // here. We also use this opportunity to set the default locale
          name: "Mr. Money",
        },
        locale: "de-DE",
      };
    },
  },
  session: {
    maxAge: 24 * 60 * 60,
  },
});

declare module "next-auth" {
  interface Session {
    /**
     * The user's locale.
     */
    locale: string;
  }
}

declare module "@auth/core/jwt" {
  interface JWT {
    /**
     * This contains the API token for our backend services.
     */
    backendAPIToken?: string;
  }
}

import { auth } from "@/lib/auth";
import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => {
  // Provide a static locale, fetch a user setting,
  // read from `cookies()`, `headers()`, etc.
  const session = await auth();
  let locale = session?.locale;

  return {
    locale,
  };
});

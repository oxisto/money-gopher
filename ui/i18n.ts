import { auth } from "@/lib/auth";
import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => {
  // Read locale from session
  const session = await auth();
  let locale = session?.locale;

  return {
    locale,
  };
});

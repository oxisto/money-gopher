import { auth } from "@/lib/auth";

export default async function Example() {
  const session = await auth();

  return <>Hello {session?.user?.name}</>;
}

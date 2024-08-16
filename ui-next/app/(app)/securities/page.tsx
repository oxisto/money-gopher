import { securitiesClient } from "@/lib/client";

export default async function Securities() {
  let client = securitiesClient(fetch)

  const securities = await client.listSecurities({}).then((res) => res.securities);
  
  return <>Securities
  
  {JSON.stringify(securities)}
  </>;
}

import { securitiesClient } from "@/lib/clients";

export default async function Securities() {
  const securities = await securitiesClient.listSecurities({}).then((res) => res.securities);
  
  return <>Securities
  
  {JSON.stringify(securities)}
  </>;
}

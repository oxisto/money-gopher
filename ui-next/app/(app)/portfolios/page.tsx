import { portfolioClient } from "@/lib/client";

export default async function Portfolios() {
  let client = portfolioClient(fetch)

  const portfolios = await client.listPortfolios({}).then((res) => res.portfolios);
  
  return <>Portfolios
  
  {JSON.stringify(portfolios)}
  </>;
}

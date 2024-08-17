import { portfolioClient } from "@/lib/clients";

export default async function Portfolios() {
  const portfolios = await portfolioClient.listPortfolios({}).then((res) => res.portfolios);
  
  return <>Portfolios
  
  {JSON.stringify(portfolios)}
  </>;
}

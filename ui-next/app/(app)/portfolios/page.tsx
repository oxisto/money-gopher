import { unstable_noStore as noStore } from 'next/cache';
import { portfolioClient } from "@/lib/clients";

export default async function Portfolios() {
  noStore();
  const portfolios = await portfolioClient.listPortfolios({}).then((res) => res.portfolios);
  
  return <>Portfolios
  
  {JSON.stringify(portfolios)}
  </>;
}

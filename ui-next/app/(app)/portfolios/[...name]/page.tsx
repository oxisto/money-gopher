import Debug from "@/components/debug";
import { unstable_noStore as noStore } from 'next/cache';
import { portfolioClient } from "@/lib/clients";

interface PortfolioProps {
  params: {
    name: string[];
  }
}

export default async function Portfolio({ params }: PortfolioProps) {
  const name = params.name.join("/");

  noStore();
  const portfolio = await portfolioClient.getPortfolio({ name: name})

  return (
    <>
      ID: <div>{portfolio.name}</div>
      Data: <Debug obj={portfolio} />
    </>
  );
}

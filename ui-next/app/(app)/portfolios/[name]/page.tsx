import Debug from "@/components/debug";
import { unstable_noStore as noStore } from "next/cache";
import { portfolioClient } from "@/lib/clients";
import PortfolioPositionsTable from "@/components/portfolio-position-table";
import Link from "next/link";

export interface PortfolioProps {
  params: {
    name: string;
  };
}

export default async function Portfolio({ params }: PortfolioProps) {
  noStore();
  const portfolio = await portfolioClient.getPortfolio({ name: params.name });
  const snapshot = await portfolioClient.getPortfolioSnapshot({
    portfolioName: params.name,
  });

  return (
    <>
      <PortfolioPositionsTable snapshot={snapshot} />

      <Link href={`/portfolios/${portfolio.name}/transactions/`}>
        Show transactions list
      </Link>

      <Link href={`/portfolios/${portfolio.name}/transactions/add`}>
        Add transaction
      </Link>
    </>
  );
}

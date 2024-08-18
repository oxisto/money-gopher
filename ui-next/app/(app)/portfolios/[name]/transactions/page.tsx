import { PortfolioProps } from "@/app/(app)/portfolios/[name]/page";
import PortfolioTransactionTable from "@/components/portfolio-transaction-table";
import { portfolioClient } from "@/lib/clients";
import { unstable_noStore as noStore } from "next/cache";

interface PortfolioTransactionProps extends PortfolioProps {}

export default async function PortfolioTransactions({
  params,
}: PortfolioTransactionProps) {
  noStore();
  const events = await portfolioClient
    .listPortfolioTransactions({ portfolioName: params.name })
    .then((res) => res.transactions);

  return <PortfolioTransactionTable events={events} />;
}

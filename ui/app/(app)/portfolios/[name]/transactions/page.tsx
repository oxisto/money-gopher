import { PortfolioProps } from "@/app/(app)/portfolios/[name]/page";
import NewPortfolioTransactionButton from "@/components/new-portfolio-transaction-button";
import PortfolioTransactionTable from "@/components/portfolio-transaction-table";
import client from "@/lib/api";

interface PortfolioTransactionProps extends PortfolioProps {}

export default async function PortfolioTransactions({
  params,
}: PortfolioTransactionProps) {
  const { data: portfolio } = await client.GET("/v1/portfolios/{name}", {
    params: { path: { name: params.name } },
  });
  const { data } = await client.GET(
    "/v1/portfolios/{portfolioName}/transactions",
    {
      params: { path: { portfolioName: params.name } },
    },
  );

  if (portfolio && data) {
    return (
      <>
        <PortfolioTransactionTable events={data?.transactions} />
        <NewPortfolioTransactionButton portfolio={portfolio} />
      </>
    );
  }
}

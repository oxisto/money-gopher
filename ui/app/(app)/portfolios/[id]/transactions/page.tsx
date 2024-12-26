import { PortfolioProps } from "@/app/(app)/portfolios/[id]/page";
import NewPortfolioTransactionButton from "@/components/new-portfolio-transaction-button";
import PortfolioTransactionTable from "@/components/portfolio-transaction-table";
import client from "@/lib/api";

interface PortfolioTransactionProps extends PortfolioProps { }

export default async function PortfolioTransactions(props: PortfolioTransactionProps) {
  const params = await props.params;
  const { data: portfolio } = await client.GET("/v1/portfolios/{id}", {
    params: { path: { id: params.id } },
  });
  const { data } = await client.GET(
    "/v1/portfolios/{portfolioId}/transactions",
    {
      params: { path: { portfolioId: params.id } },
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

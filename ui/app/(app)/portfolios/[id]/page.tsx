import Button from "@/components/button";
import NewPortfolioTransactionButton from "@/components/new-portfolio-transaction-button";
import PortfolioPositionsTable from "@/components/portfolio-position-table";
import client from "@/lib/api";
import Link from "next/link";

export interface PortfolioProps {
  params: Promise<{
    id: string;
  }>;
}

export default async function Portfolio(props: PortfolioProps) {
  const params = await props.params;
  const { data: portfolio } = await client.GET("/v1/portfolios/{id}", {
    params: { path: { id: params.id } },
  });
  const { data: snapshot } = await client.GET(
    "/v1/portfolios/{portfolioId}/snapshot",
    {
      params: { path: { portfolioId: params.id } },
    }
  );

  if (portfolio && snapshot) {
    return (
      <>
        <PortfolioPositionsTable snapshot={snapshot} />

        <div className="space-x-2">
          <Link href={`/portfolios/${portfolio.id}/transactions/`}>
            <Button>Show transactions list</Button>
          </Link>

          <NewPortfolioTransactionButton portfolio={portfolio} />
        </div>
      </>
    );
  }
}

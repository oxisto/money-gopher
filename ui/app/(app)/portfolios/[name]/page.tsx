import Button from "@/components/button";
import NewPortfolioTransactionButton from "@/components/new-portfolio-transaction-button";
import PortfolioPositionsTable from "@/components/portfolio-position-table";
import client from "@/lib/api";
import Link from "next/link";

export interface PortfolioProps {
  params: Promise<{
    name: string;
  }>;
}

export default async function Portfolio(props: PortfolioProps) {
  const params = await props.params;
  const { data: portfolio } = await client.GET("/v1/portfolios/{name}", {
    params: { path: { name: params.name } },
  });
  const { data: snapshot } = await client.GET(
    "/v1/portfolios/{portfolioName}/snapshot",
    {
      params: { path: { portfolioName: params.name } },
    }
  );

  if (portfolio && snapshot) {
    return (
      <>
        <PortfolioPositionsTable snapshot={snapshot} />

        <div className="space-x-2">
          <Link href={`/portfolios/${portfolio.name}/transactions/`}>
            <Button>Show transactions list</Button>
          </Link>

          <NewPortfolioTransactionButton portfolio={portfolio} />
        </div>
      </>
    );
  }
}

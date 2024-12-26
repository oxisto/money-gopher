import Button from "@/components/button";
import { SchemaPortfolio } from "@/lib/api";
import Link from "next/link";

interface NewPortfolioTransactionButtonProps {
  portfolio: SchemaPortfolio;
}

export default function NewPortfolioTransactionButton({
  portfolio,
}: NewPortfolioTransactionButtonProps) {
  return (
    <Link href={`/portfolios/${portfolio.id}/transactions/new`}>
      <Button>New transaction</Button>
    </Link>
  );
}

import Button from "@/components/button";
import { Portfolio } from "@/lib/api";
import Link from "next/link";

interface NewPortfolioTransactionButtonProps {
  portfolio: Portfolio;
}

export default function NewPortfolioTransactionButton({
  portfolio,
}: NewPortfolioTransactionButtonProps) {
  return (
    <Link href={`/portfolios/${portfolio.name}/transactions/new`}>
      <Button>New transaction</Button>
    </Link>
  );
}

import { modifyTransaction } from "@/actions/modify-transaction";
import EditPortfolioTransactionForm from "@/components/edit-portfolio-transaction";
import client from "@/lib/api";

interface PortfolioTransactionProps {
  params: Promise<{
    transactionId: string;
    id: string;
  }>;
}
interface EditPortfolioTransactionProps extends PortfolioTransactionProps { }

export default async function EditPortfolioTransaction(props: EditPortfolioTransactionProps) {
  const params = await props.params;
  const { data, error } = await client.GET("/v1/securities");
  const create = params.transactionId == "new";
  if (create && data) {
    return <EditPortfolioTransactionForm
      action={modifyTransaction}
      create={true}
      event={{
        id: "new",
        time: new Date().toISOString(),
        portfolioId: params.id,
        securityId: "",
        amount: 1,
        type: "PORTFOLIO_EVENT_TYPE_BUY",
        price: { value: 0, symbol: "EUR" },
        fees: { value: 0, symbol: "EUR" },
        taxes: { value: 0, symbol: "EUR" },
      }}
      securities={data?.securities}
    />;
  } else {
    const { data: event } = await client.GET("/v1/transactions/{id}", {
      params: { path: { id: params.transactionId } },
    });

    if (event && data) {
      return (
        <EditPortfolioTransactionForm
          action={modifyTransaction}
          create={create}
          securities={data?.securities}
          event={event}
        />
      );
    } else {
      return <>Transaction not found</>;
    }
  }
}

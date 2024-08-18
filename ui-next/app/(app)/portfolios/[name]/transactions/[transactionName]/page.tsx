import EditPortfolioTransactionForm from "@/components/edit-portfolio-transaction";
import client, { PortfolioEvent } from "@/lib/api";

interface PortfolioTransactionProps {
  params: {
    transactionName: string;
    name: string;
  };
}
interface EditPortfolioTransactionProps extends PortfolioTransactionProps {}

export default async function EditPortfolioTransaction({
  params,
}: EditPortfolioTransactionProps) {
  async function editTransaction(formData: FormData) {
    "use server";
    console.log(formData);
  }

  const create = params.transactionName == "new";

  if (create) {
    <EditPortfolioTransactionForm
      action={editTransaction}
      create={true}
      event={{
        name: "",
        time: new Date().toISOString(),
        portfolioName: params.name,
        securityName: "",
        amount: 1,
        type: "PORTFOLIO_EVENT_TYPE_BUY",
        price: { value: 0, symbol: "EUR" },
        fees: { value: 0, symbol: "EUR" },
        taxes: { value: 0, symbol: "EUR" },
      }}
    />;
  } else {
    const { data: event } = await client.GET("/v1/transactions/{name}", {
      params: { path: { name: params.transactionName } },
    });

    if (event) {
      return (
        <EditPortfolioTransactionForm
          action={editTransaction}
          create={create}
          event={event}
        />
      );
    } else {
      return <>Transaction not found</>;
    }
  }
}

import Button from "@/components/button";
import { EditPortfolioTransactionFormInner } from "@/components/edit-portfolio-transaction-inner";
import client, { PortfolioEvent } from "@/lib/api";

interface EditPortfolioTransactionFormProps {
  create: Boolean;
  event: PortfolioEvent;
  action: (formData: FormData) => void;
}

export default async function EditPortfolioTransactionForm({
  create = false,
  event,
  action,
}: EditPortfolioTransactionFormProps) {
  const { data } = await client.GET("/v1/securities");
  const securityOptions =
    data?.securities.map((s) => {
      return { value: s.name, display: s.displayName };
    }) ?? [];

  return (
    <form action={action}>
      <EditPortfolioTransactionFormInner
        create={create}
        initial={event}
        securityOptions={securityOptions}
      />
      <Button type="submit">{create ? "Create" : "Save"}</Button>
    </form>
  );
}

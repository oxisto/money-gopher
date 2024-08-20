import client, { PortfolioEvent } from "@/lib/api";
import { dateTimeLocalFormat } from "@/lib/util";
import dayjs from "dayjs";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

export async function modifyTransaction(formData: FormData) {
  "use server";

  // Build a portfolio transaction from the formdata
  let event: PortfolioEvent = formDataToPortfolioEvent(formData);
  const newEvent =
    event.name == "new"
      ? await createTransaction(event)
      : await editTransaction(event);
  revalidatePath(`/portfolios/${newEvent.portfolioName}/transactions`);
  redirect(`/portfolios/${newEvent.portfolioName}/transactions`);
}

async function createTransaction(
  event: PortfolioEvent,
): Promise<PortfolioEvent> {
  const { data: newEvent, error } = await client.POST(
    "/v1/portfolios/{transaction.portfolio_name}/transactions",
    {
      params: {
        path: {
          "transaction.portfolio_name": event.portfolioName,
        },
      },
      body: event,
    },
  );
  if (error != null) {
    throw error;
  }
  return newEvent;
}

async function editTransaction(event: PortfolioEvent): Promise<PortfolioEvent> {
  const { data: newEvent, error } = await client.PUT(
    "/v1/transactions/{transaction.name}",
    {
      params: {
        path: {
          "transaction.name": event.name,
        },
        query: {
          updateMask: "amount,price,fees,taxes,securityName,time",
        },
      },
      body: event,
    },
  );
  if (error != null) {
    throw error;
  }
  return newEvent;
}

function formDataToPortfolioEvent(formData: FormData): PortfolioEvent {
  return {
    name: formData.get("name")?.toString() ?? "",
    portfolioName: formData.get("portfolioName")?.toString() ?? "",
    securityName: formData.get("securityName[value]")?.toString() ?? "",
    type:
      (formData.get("type[value]")?.toString() as
        | "PORTFOLIO_EVENT_TYPE_UNSPECIFIED"
        | "PORTFOLIO_EVENT_TYPE_BUY"
        | "PORTFOLIO_EVENT_TYPE_SELL"
        | "PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND"
        | "PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND"
        | "PORTFOLIO_EVENT_TYPE_DIVIDEND"
        | "PORTFOLIO_EVENT_TYPE_INTEREST"
        | "PORTFOLIO_EVENT_TYPE_DEPOSIT_CASH"
        | "PORTFOLIO_EVENT_TYPE_WITHDRAW_CASH"
        | "PORTFOLIO_EVENT_TYPE_ACCOUNT_FEES"
        | "PORTFOLIO_EVENT_TYPE_TAX_REFUND") ?? "",
    time: dayjs(
      formData.get("time")?.toString() ?? "",
      dateTimeLocalFormat,
    ).toISOString(),
    amount: Number(formData.get("amount")),
    price: {
      value: Number(formData.get("price.value")) * 100,
      symbol: formData.get("price.symbol")?.toString() ?? "",
    },
    fees: {
      value: Number(formData.get("fees.value")) * 100,
      symbol: formData.get("fees.symbol")?.toString() ?? "",
    },
    taxes: {
      value: Number(formData.get("taxes.value")) * 100,
      symbol: formData.get("taxes.symbol")?.toString() ?? "",
    },
  };
}

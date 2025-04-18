"use client";

import Button from "@/components/button";
import FormattedCurrency from "@/components/formatted-currency";
import CurrencyInput from "@/components/forms/currency-input";
import DateInput from "@/components/forms/date-input";
import ListBox from "@/components/forms/listbox";
import { SchemaCurrency, SchemaPortfolioEvent, SchemaSecurity } from "@/lib/api";
import { useMemo, useState } from "react";

interface EditPortfolioTransactionFormProps {
  create: Boolean;
  event: SchemaPortfolioEvent;
  securities: SchemaSecurity[];
  action: (formData: FormData) => void;
}

export default function EditPortfolioTransactionForm({
  create = false,
  event,
  securities,
  action,
}: EditPortfolioTransactionFormProps) {
  let [data, setData] = useState(event);
  let isSecurityTransaction = useMemo(
    () =>
      data.type == "PORTFOLIO_EVENT_TYPE_BUY" ||
      data.type == "PORTFOLIO_EVENT_TYPE_SELL",
    [data],
  );
  const total = useMemo<SchemaCurrency>(() => {
    return {
      value:
        data.amount * (data.price.value ?? 0) +
        (data.fees.value ?? 0) +
        (data.taxes?.value ?? 0),
      symbol: "EUR",
    };
  }, [data]);

  type typeType = typeof data.type;

  const typeOptions: { value: typeType; display: string }[] = [
    { value: "PORTFOLIO_EVENT_TYPE_BUY", display: "Buy" },
    { value: "PORTFOLIO_EVENT_TYPE_SELL", display: "Sell" },
    {
      value: "PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND",
      display: "Delivery Inbound",
    },
    {
      value: "PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND",
      display: "Delibery Outbound",
    },
    { value: "PORTFOLIO_EVENT_TYPE_DIVIDEND", display: "Dividend" },
    { value: "PORTFOLIO_EVENT_TYPE_INTEREST", display: "Interest" },
    { value: "PORTFOLIO_EVENT_TYPE_DEPOSIT_CASH", display: "Deposit Cash" },
    { value: "PORTFOLIO_EVENT_TYPE_WITHDRAW_CASH", display: "Withdraw Cash" },
    { value: "PORTFOLIO_EVENT_TYPE_ACCOUNT_FEES", display: "Account Fees" },
    { value: "PORTFOLIO_EVENT_TYPE_TAX_REFUND", display: "Tax Refund" },
  ];
  const securityOptions = securities.map((s) => {
    return { value: s.id, display: s.displayName };
  });

  return (
    <form action={action}>
      <input type="hidden" name="id" value={data.id} />
      <input type="hidden" name="portfolioId" value={data.portfolioId} />
      <div className="space-y-12 sm:space-y-16">
        <div>
          <h2 className="text-base font-semibold leading-7 text-gray-900">
            {create ? <>Create Transaction</> : <>Edit Transaction</>}
          </h2>
          <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-600">
            {create ? (
              <>
                This allows you to create a new transaction and add it to
                portfolio <b>{data.portfolioId}</b>.
              </>
            ) : (
              <>
                This allows you to edit the existing transaction{" "}
                <b>{data.id}</b> in portfolio <b>{data.portfolioId}</b>.
              </>
            )}
          </p>
          <div
            className="mt-10 space-y-8 border-b border-gray-900/10 pb-12 sm:space-y-0 sm:divide-y
              sm:divide-gray-900/10 sm:border-t sm:pb-0"
          >
            <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
              <label
                htmlFor="username"
                className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
              >
                Transaction Type
              </label>
              <div className="mt-2 sm:col-span-2 sm:mt-0">
                <ListBox
                  name="type"
                  value={data.type}
                  options={typeOptions}
                  onChange={(value) => setData({ ...data, type: value })}
                />
              </div>
            </div>

            <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
              <label
                htmlFor="username"
                className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
              >
                Date
              </label>
              <div className="mt-2 sm:col-span-2 sm:mt-0">
                <DateInput
                  name="time"
                  value={data.time}
                  onChange={(value) => setData({ ...data, time: value })}
                />
              </div>
            </div>

            {isSecurityTransaction && (
              <>
                <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
                  <label
                    htmlFor="username"
                    className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
                  >
                    Security
                  </label>
                  <div className="mt-2 sm:col-span-2 sm:mt-0">
                    <ListBox
                      name="securityId"
                      value={data.securityId}
                      options={securityOptions}
                      onChange={(value) =>
                        setData({ ...data, securityId: value })
                      }
                    />
                  </div>
                </div>

                <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
                  <label
                    htmlFor="amount"
                    className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
                  >
                    Amount
                  </label>
                  <div className="mt-2 sm:col-span-2 sm:mt-0">
                    <input
                      type="number"
                      name="amount"
                      id="amount"
                      min="1"
                      step="any"
                      placeholder="1"
                      value={data.amount}
                      onChange={(e) => {
                        setData({ ...data, amount: e.target.valueAsNumber });
                      }}
                      className="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1
                        ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset
                        focus:ring-indigo-600 sm:max-w-xs sm:text-sm sm:leading-6"
                    />
                  </div>
                </div>
              </>
            )}

            <CurrencyInput
              name="price"
              value={data.price}
              symbol="EUR"
              onChange={(value) => {
                setData({ ...data, price: value });
              }}
            >
              Price
            </CurrencyInput>
            {isSecurityTransaction && (
              <>
                <CurrencyInput
                  name="fees"
                  value={data.fees}
                  symbol="EUR"
                  onChange={(value) => {
                    setData({ ...data, fees: value });
                  }}
                >
                  Fees
                </CurrencyInput>
                <CurrencyInput
                  name="taxes"
                  value={data.taxes}
                  symbol="EUR"
                  onChange={(value) => {
                    setData({ ...data, taxes: value });
                  }}
                >
                  Taxes
                </CurrencyInput>
              </>
            )}

            <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
              <div className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
                Total
              </div>
              <div className="mt-2 sm:col-span-2 sm:mt-0">
                <div className="block w-full text-gray-900 sm:max-w-xs sm:text-sm sm:leading-6">
                  <FormattedCurrency value={total} />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <Button type="submit">{create ? "Create" : "Save"}</Button>
    </form>
  );
}

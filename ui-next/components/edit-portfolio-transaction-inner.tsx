"use client";

import { currencyValue } from "@/lib/util";
import { useState, useEffect, useMemo } from "react";
import CurrencyInput from "@/components/forms/currency-input";
import ListBox from "@/components/forms/listbox";
import { PortfolioEvent } from "@/lib/api";

interface EditPortfolioTransactionFormInnerProps {
  create: Boolean;
  initial: PortfolioEvent;
  securityOptions: { value: any; display: string }[];
}

export function EditPortfolioTransactionFormInner({
  create,
  initial,
  securityOptions,
}: EditPortfolioTransactionFormInnerProps) {
  let [data, setData] = useState(initial);
  let isSecurityTransaction = useMemo(
    () =>
      data.type == "PORTFOLIO_EVENT_TYPE_BUY" ||
      data.type == "PORTFOLIO_EVENT_TYPE_SELL",
    [data]
  );
  const total = useMemo(
    () => data.amount * (data.price.value + data.fees.value + data.taxes.value),
    [data]
  );
  console.log(total);

  return (
    <div className="space-y-12 sm:space-y-16">
      <div>
        <h2 className="text-base font-semibold leading-7 text-gray-900">
          {create ? <>Create Transaction</> : <>Edit Transaction</>}
        </h2>
        <p className="mt-1 max-w-2xl text-sm leading-6 text-gray-600">
          {create ? (
            <>
              This allows you to create a new transaction and add it to
              portfolio <b>{data.portfolioName}</b>.
            </>
          ) : (
            <>
              This allows you to edit the existing transaction{" "}
              <b>{data.name}</b> in portfolio <b>{data.portfolioName}</b>.
            </>
          )}
        </p>
        <div className="mt-10 space-y-8 border-b border-gray-900/10 pb-12 sm:space-y-0 sm:divide-y sm:divide-gray-900/10 sm:border-t sm:pb-0">
          <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
            <label
              htmlFor="username"
              className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
            >
              Transaction Type
            </label>
            <div className="mt-2 sm:col-span-2 sm:mt-0">
              <ListBox
                value={data.type}
                options={[]}
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
            <div className="mt-2 sm:col-span-2 sm:mt-0"></div>
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
                    value={data.securityName}
                    options={securityOptions}
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
                    defaultValue={data.amount}
                    className="
          block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300
          placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:max-w-xs sm:text-sm sm:leading-6"
                  />
                </div>
              </div>
            </>
          )}

          <CurrencyInput name="price" value={data.price} symbol="EUR">
            Price
          </CurrencyInput>
          {isSecurityTransaction && (
            <>
              <CurrencyInput name="fees" value={data.fees} symbol="EUR">
                Fees
              </CurrencyInput>
              <CurrencyInput name="taxes" value={data.taxes} symbol="EUR">
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
                {currencyValue(total, "EUR")}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

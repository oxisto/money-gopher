import { Currency } from "@/lib/api";
import { toDecimal } from "@/lib/currency";
import { Field, Input, Label } from "@headlessui/react";
import { useFormatter } from "next-intl";
import { ChangeEvent, InputHTMLAttributes, useState } from "react";

interface CurrencyInputProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "value" | "onChange"> {
  name: string;
  value?: Currency;
  symbol: string;
  children: React.ReactNode;
  onChange?: (value: Currency) => void;
}

export default function CurrencyInput({
  name,
  value = { value: 0, symbol: "EUR" },
  children,
  onChange,
  ...rest
}: CurrencyInputProps) {
  const format = useFormatter();
  let [internal, setInternal] = useState(toDecimal(value));

  function onCurrencyChange(e: ChangeEvent<HTMLInputElement>) {
    // update our internal value...
    setInternal(e.target.valueAsNumber);

    // ... and propagate changes back to parent
    onChange?.call(null, {
      value: e.target.valueAsNumber * 100,
      symbol: value.symbol,
    });
  }

  return (
    <Field className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
      <Label className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5">
        {children}
      </Label>
      <div className="relative mt-2 rounded-md shadow-sm">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <span className="text-gray-500 sm:text-sm">€</span>
        </div>
        <Input
          type="number"
          name={`${name}.value`}
          value={internal}
          className="block w-full rounded-md border-0 py-1.5 pl-7 pr-12 text-gray-900 ring-1
            ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset
            focus:ring-indigo-600 sm:text-sm sm:leading-6"
          placeholder={format.number(0.0, {
            maximumFractionDigits: 2,
            minimumFractionDigits: 2,
          })}
          aria-describedby="price-currency"
          onChange={onCurrencyChange}
          min="0"
          step="any"
          {...rest}
        />
        <input
          type="hidden"
          name={`${name}.symbol`}
          defaultValue={value.symbol}
        />
        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
          <span className="text-gray-500 sm:text-sm" id="price-currency">
            {value.symbol}
          </span>
        </div>
      </div>
    </Field>
  );
}

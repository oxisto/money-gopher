import { Currency } from "@/lib/api";

interface CurrencyInputProps {
  name: string;
  value: Currency;
  symbol: string;
  children: React.ReactNode;
}

export default function CurrencyInput({
  name,
  value,
  children,
}: CurrencyInputProps) {
  return (
    <div className="sm:grid sm:grid-cols-3 sm:items-start sm:gap-4 sm:py-6">
      <label
        htmlFor={name}
        className="block text-sm font-medium leading-6 text-gray-900 sm:pt-1.5"
      >
        {children}
      </label>
      <div className="relative mt-2 rounded-md shadow-sm">
        <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
          <span className="text-gray-500 sm:text-sm">€</span>
        </div>
        <input
          type="number"
          name={name}
          id={name}
          defaultValue={value.value}
          className="block w-full rounded-md border-0 py-1.5 pl-7 pr-12 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400
  focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
          placeholder="0.00"
          aria-describedby="price-currency"
        />
        <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
          <span className="text-gray-500 sm:text-sm" id="price-currency">
            {value.symbol}
          </span>
        </div>
      </div>
    </div>
  );
}

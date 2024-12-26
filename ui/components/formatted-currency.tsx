import { SchemaCurrency } from "@/lib/api";
import { toDecimal } from "@/lib/currency";
import { useFormatter } from "next-intl";

interface FormattedCurrencyProps {
  value: SchemaCurrency;
}

export default function FormattedCurrency({ value }: FormattedCurrencyProps) {
  const format = useFormatter();

  return format.number(toDecimal(value), {
    style: "currency",
    currency: value.symbol,
  });
}

import { Currency } from "@/lib/api";
import { useFormatter } from "next-intl";

interface FormattedCurrencyProps {
  currency: Currency;
}

export default function FormattedCurrency({
  currency,
}: FormattedCurrencyProps) {
  const format = useFormatter();

  return format.number(currency.value / 100, {
    style: "currency",
    currency: currency.symbol,
  });
}

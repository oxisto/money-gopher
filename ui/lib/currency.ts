import { SchemaCurrency } from "@/lib/api";

/**
 * This function returns a currency value to a decimal. Internally,
 * {@link Currency} uses an integer-based system on the lowest currency unit
 * (e.g. cents on EUR) in order to avoid floating values when calculating.
 * However, when we want to display the currency, we want to display decimals of
 * the higher unit (e.g. EUR).
 *
 * @param currency the currency to convert
 * @returns a decimal representation
 */
export function toDecimal(currency?: SchemaCurrency): number {
  if (
    currency == undefined ||
    currency.value == undefined ||
    currency.value == 0
  ) {
    return 0;
  } else {
    return currency.value / 100;
  }
}

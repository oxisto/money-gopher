import { useFormatter } from "next-intl";

interface FormattedPercentageProps {
  /**
   * The fractional number representing a percentage.
   */
  value: number;
}
export default function FormattedPercentage({
  value,
}: FormattedPercentageProps) {
  const format = useFormatter();

  return format.number(value, {
    style: "percent",
    maximumFractionDigits: 2,
  });
}

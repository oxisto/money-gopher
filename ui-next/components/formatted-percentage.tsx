interface FormattedPercentageProps {
  /**
   * The fractional number representing a percentage.
   */
  number: number;
}
export default function FormattedPercentage({
  number,
}: FormattedPercentageProps) {
  return (
    <>
      {Intl.NumberFormat(navigator.language, {
        maximumFractionDigits: 2,
      }).format(number * 100)}{" "}
      %
    </>
  );
}

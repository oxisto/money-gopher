import { components, SchemaPortfolioPosition } from "@/lib/api/v1";
import { classNames, shorten } from "@/lib/util";
import {
  ArrowDownIcon,
  ArrowRightIcon,
  ArrowUpIcon,
} from "@heroicons/react/24/outline";
import FormattedCurrency from "./formatted-currency";
import FormattedPercentage from "./formatted-percentage";

interface PortfolioPositionRowProps {
  position: SchemaPortfolioPosition;
}

export default function PortfolioPositionRow({
  position,
}: PortfolioPositionRowProps) {
  const Icon =
    Math.abs(position.gains) < 0.01
      ? ArrowRightIcon
      : position.gains < 0
        ? ArrowDownIcon
        : ArrowUpIcon;

  if (position.security) {
    return (
      <tr>
        <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
          <div className="text-gray-900">
            {shorten(position.security.displayName)}
          </div>
          <div className="text-gray-400">{position.security.id}</div>
        </td>
        <td
          className="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500
            md:table-cell"
        >
          {Intl.NumberFormat(navigator.language, {
            maximumFractionDigits: 2,
          }).format(position.amount)}
        </td>
        <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
          <div className="text-gray-500">
            <FormattedCurrency value={position.purchasePrice} />
          </div>
          <div className="text-gray-400">
            <FormattedCurrency value={position.purchaseValue} />
          </div>
        </td>
        <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
          <div className="text-gray-500">
            <FormattedCurrency value={position.marketPrice} />
          </div>
          <div className="text-gray-400">
            <FormattedCurrency value={position.marketValue} />
          </div>
        </td>
        <td
          className={classNames(
            Math.abs(position.gains) <= 0.01
              ? "text-gray-500"
              : position.gains < 0
                ? "text-red-500"
                : "text-green-500",
            "whitespace-nowrap px-3 py-2 text-right text-sm",
          )}
        >
          <div>
            <FormattedPercentage value={position.gains} />{" "}
            <Icon className="float-right mt-0.5 h-4 w-4" aria-hidden="true" />
          </div>
          <div className="pr-4">
            <FormattedCurrency value={position.profitOrLoss} />
          </div>
        </td>
      </tr>
    );
  }
}

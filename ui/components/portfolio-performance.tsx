import { SchemaPortfolioSnapshot } from "@/lib/api";
import { classNames } from "@/lib/util";
import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
} from "@heroicons/react/24/outline";
import FormattedCurrency from "./formatted-currency";
import FormattedPercentage from "./formatted-percentage";

interface PortfolioPerformanceProps {
  snapshot: SchemaPortfolioSnapshot;
  showIcon: boolean;
}

export default function PortfolioPerformance({
  snapshot,
  showIcon,
}: PortfolioPerformanceProps) {
  const Icon =
    snapshot.totalGains > 0 ? ArrowTrendingUpIcon : ArrowTrendingDownIcon;
  return (
    <div
      className={classNames(
        snapshot.totalGains < 0 ? "text-red-400" : "text-green-400",
        "flex items-center",
      )}
    >
      {showIcon && (
        <Icon
          className={classNames(
            snapshot.totalGains < 0 ? "text-red-400" : "text-green-400",
            "mr-1.5 h-5 w-5 flex-shrink-0",
          )}
          aria-hidden="true"
        />
      )}
      <FormattedPercentage value={snapshot.totalGains} />{" "}
      <FormattedCurrency value={snapshot.totalProfitOrLoss} />
    </div>
  );
}

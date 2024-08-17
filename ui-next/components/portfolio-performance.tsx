import { PortfolioSnapshot } from "@/lib/gen/mgo_pb";
import { classNames, currency } from "@/lib/util";
import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
} from "@heroicons/react/24/outline";
import FormattedPercentage from "./formatted-percentage";

interface PortfolioPerformanceProps {
  snapshot: PortfolioSnapshot;
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
        "flex items-center"
      )}
    >
      {showIcon && (
        <Icon
          className={classNames(
            snapshot.totalGains < 0 ? "text-red-400" : "text-green-400",
            "mr-1.5 h-5 w-5 flex-shrink-0"
          )}
          aria-hidden="true"
        />
      )}
      <FormattedPercentage number={snapshot.totalGains} />&nbsp;({currency(snapshot.totalProfitOrLoss)})
    </div>
  );
}

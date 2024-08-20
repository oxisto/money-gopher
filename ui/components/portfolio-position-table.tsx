import FormattedCurrency from "@/components//formatted-currency";
import PortfolioPositionRow from "@/components/portfolio-position-row";
import TableSorter from "@/components/table-sorter";
import { PortfolioPosition, PortfolioSnapshot } from "@/lib/api";
import { classNames } from "@/lib/util";
import {
  ArrowDownIcon,
  ArrowRightIcon,
  ArrowUpIcon,
} from "@heroicons/react/24/outline";
import FormattedPercentage from "./formatted-percentage";

const sorters = new Map<
  string,
  (a: PortfolioPosition, b: PortfolioPosition) => number
>();
sorters.set("displayName", (a: PortfolioPosition, b: PortfolioPosition) => {
  return (
    a.security?.displayName.localeCompare(b.security?.displayName ?? "") ?? 0
  );
});
sorters.set("amount", (a: PortfolioPosition, b: PortfolioPosition) => {
  return a.amount - b.amount;
});
sorters.set("purchaseValue", (a: PortfolioPosition, b: PortfolioPosition) => {
  return (a.purchaseValue?.value ?? 0) - (b.purchaseValue?.value ?? 0);
});
sorters.set("marketValue", (a: PortfolioPosition, b: PortfolioPosition) => {
  return (a.marketValue?.value ?? 0) - (b.marketValue?.value ?? 0);
});

interface PortfolioPositionsTableProps {
  snapshot: PortfolioSnapshot;
}

function getPositions(
  snapshot: PortfolioSnapshot,
  sortBy: string,
  asc: boolean,
): PortfolioPosition[] {
  let positions = Object.values(snapshot.positions ?? {});
  return positions.sort((a: PortfolioPosition, b: PortfolioPosition) => {
    const sort = sorters.get(sortBy)?.call(null, a, b) ?? 0;
    return asc ? sort : -sort;
  });
}

export default function PortfolioPositionsTable({
  snapshot,
}: PortfolioPositionsTableProps) {
  //let [sortBy, changeSortBy] = useState("displayName");
  //let [asc, setAsc] = useState(true);
  let sortBy = "displayName";
  let asc = false;

  const Icon =
    snapshot.totalGains < 0
      ? ArrowDownIcon
      : snapshot.totalGains < 0.01
        ? ArrowRightIcon
        : ArrowUpIcon;

  function toggleSortDirection() {
    // nothing yet
  }

  function changeSortBy(column: string) {
    // nothing yet
  }

  const positions = getPositions(snapshot, sortBy, asc);
  return (
    <div className="-mx-4 mt-8 sm:-mx-0">
      <table className="min-w-full divide-y divide-gray-300">
        <thead>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              <TableSorter
                active={sortBy == "displayName"}
                column="displayName"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={(column) => changeSortBy(column)}
              >
                Name
              </TableSorter>
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"
            >
              <TableSorter
                active={sortBy == "amount"}
                column="amount"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={(column) => changeSortBy("amount")}
              >
                Amount
              </TableSorter>
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
            >
              <TableSorter
                active={sortBy == "purchaseValue"}
                column="purchaseValue"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={(column) => changeSortBy("purchaseValue")}
              >
                Purchase Value
              </TableSorter>
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
            >
              <TableSorter
                active={sortBy == "marketValue"}
                column="marketValue"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={(column) => changeSortBy("marketValue")}
              >
                Market Value
              </TableSorter>
            </th>
            <th
              scope="col"
              className="px-3 py-3.5 text-right text-sm font-semibold text-gray-900"
            >
              <TableSorter
                active={sortBy == "profit"}
                column="profit"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={(column) => changeSortBy("profit")}
              >
                Profit/Loss
              </TableSorter>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {positions.map((position, idx) => (
            <PortfolioPositionRow key={idx} position={position} />
          ))}
        </tbody>
        <tfoot>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              Total Assets
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 md:table-cell"
            ></th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell"
            >
              <FormattedCurrency value={snapshot.totalPurchaseValue} />
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
            >
              <FormattedCurrency value={snapshot.totalMarketValue} />
            </th>
            <th
              scope="col"
              className={classNames(
                snapshot.totalGains < 0
                  ? "text-red-500"
                  : snapshot.totalGains <= 0.01
                    ? "text-gray-500"
                    : "text-green-500",
                "px-3 py-3.5 text-right text-sm font-semibold",
              )}
            >
              <div>
                <FormattedPercentage value={snapshot.totalGains} />
                <Icon
                  className="float-right mt-0.5 h-4 w-4"
                  aria-hidden="true"
                />
              </div>
              <div className="pr-4">
                <FormattedCurrency value={snapshot.totalProfitOrLoss} />
              </div>
            </th>
          </tr>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              Cash Value
            </th>
            <th></th>
            <th></th>
            <th
              scope="col"
              className={classNames(
                snapshot.cash?.value < 0 ? "text-red-500" : "",
                "px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell",
              )}
            >
              <FormattedCurrency value={snapshot.cash} />
            </th>
          </tr>
          <tr>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              Total Portfolio Value
            </th>
            <th></th>
            <th></th>
            <th
              scope="col"
              className={classNames(
                snapshot.cash?.value < 0 ? "text-red-500" : "",
                "px-3 py-3.5 text-right text-sm font-semibold text-gray-900 lg:table-cell",
              )}
            >
              <FormattedCurrency value={snapshot.totalPortfolioValue} />
            </th>
          </tr>
        </tfoot>
      </table>
    </div>
  );
}

import PortfolioTransactionRow from "@/components/portfolio-transaction-row";
import TableSorter from "@/components//table-sorter";
import { PortfolioEvent, PortfolioSnapshot } from "@/lib/gen/mgo_pb";

interface PortfolioTransactionTableProps {
  events: PortfolioEvent[];
}

let sortBy = "name";
let asc = false;

const sorters = new Map<
  string,
  (a: PortfolioEvent, b: PortfolioEvent) => number
>();
sorters.set("time", (a: PortfolioEvent, b: PortfolioEvent) => {
  return (a.time?.toDate() ?? 0) < (b.time?.toDate() ?? 0) ? -1 : 1;
});
sorters.set("securityName", (a: PortfolioEvent, b: PortfolioEvent) => {
  return a.securityName.localeCompare(b.securityName);
});
sorters.set("amount", (a: PortfolioEvent, b: PortfolioEvent) => {
  return a.amount - b.amount;
});
sorters.set("price", (a: PortfolioEvent, b: PortfolioEvent) => {
  return (a.price?.value ?? 0) - (b.price?.value ?? 0);
});
sorters.set("fees", (a: PortfolioEvent, b: PortfolioEvent) => {
  return (a.fees?.value ?? 0) - (b.fees?.value ?? 0);
});
sorters.set("taxes", (a: PortfolioEvent, b: PortfolioEvent) => {
  return (a.taxes?.value ?? 0) - (b.taxes?.value ?? 0);
});

function getPositions(
  transactions: PortfolioEvent[],
  sortBy: string,
  asc: boolean
): PortfolioEvent[] {
  return transactions.sort((a: PortfolioEvent, b: PortfolioEvent) => {
    const sort = sorters.get(sortBy)?.call(null, a, b) ?? 0;
    return asc ? sort : -sort;
  });
}

function toggleSortDirection() {
  asc = !asc;
}

function changeSortBy(column: string) {
  sortBy = column;
}

export default function PortfolioTransactionTable({
  events,
}: PortfolioTransactionTableProps) {
  const sorted = getPositions(events, sortBy, asc);

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
                active={sortBy == "date"}
                column="date"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={() => changeSortBy("date")}
              >
                Date
              </TableSorter>
            </th>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              <TableSorter
                active={sortBy == "date"}
                column="date"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={() => changeSortBy("date")}
              >
                Type
              </TableSorter>
            </th>
            <th
              scope="col"
              className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0"
            >
              <TableSorter
                active={sortBy == "displayName"}
                column="displayName"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={() => changeSortBy("displayName")}
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
                onChangeSortBy={() => changeSortBy("amount")}
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
                onChangeSortBy={() => changeSortBy("purchaseValue")}
              >
                Price
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
                onChangeSortBy={() => changeSortBy("marketValue")}
              >
                Fees
              </TableSorter>
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
            >
              <TableSorter
                active={sortBy == "profit"}
                column="profit"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={() => changeSortBy("profit")}
              >
                Taxes
              </TableSorter>
            </th>
            <th
              scope="col"
              className="hidden px-3 py-3.5 text-right text-sm font-semibold text-gray-900 sm:table-cell"
            >
              <TableSorter
                active={sortBy == "total"}
                column="total"
                onChangeDirection={toggleSortDirection}
                onChangeSortBy={() => changeSortBy("total")}
              >
                Total
              </TableSorter>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {sorted.map((event, idx) => (
            <PortfolioTransactionRow event={event} key={idx} currency="EUR" />)
          )}
        </tbody>
      </table>
    </div>
  );
}

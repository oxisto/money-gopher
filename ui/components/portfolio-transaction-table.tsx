import TableSorter from "@/components//table-sorter";
import PortfolioTransactionRow from "@/components/portfolio-transaction-row";
import client, { SchemaPortfolioEvent } from "@/lib/api";

interface PortfolioTransactionTableProps {
  events: SchemaPortfolioEvent[];
}

let sortBy = "id";
let asc = false;

const sorters = new Map<
  string,
  (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => number
>();
sorters.set("time", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return (a.time ?? 0) < (b.time ?? 0) ? -1 : 1;
});
sorters.set("securityId", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return a.securityId.localeCompare(b.securityId);
});
sorters.set("amount", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return a.amount - b.amount;
});
sorters.set("price", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return (a.price?.value ?? 0) - (b.price?.value ?? 0);
});
sorters.set("fees", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return (a.fees?.value ?? 0) - (b.fees?.value ?? 0);
});
sorters.set("taxes", (a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
  return (a.taxes?.value ?? 0) - (b.taxes?.value ?? 0);
});

function getPositions(
  transactions: SchemaPortfolioEvent[],
  sortBy: string,
  asc: boolean
): SchemaPortfolioEvent[] {
  return transactions?.sort((a: SchemaPortfolioEvent, b: SchemaPortfolioEvent) => {
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

export default async function PortfolioTransactionTable({
  events,
}: PortfolioTransactionTableProps) {
  const { data } = await client.GET("/v1/securities")
  const securities = data?.securities ?? []
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
          {sorted?.map((event, idx) => (
            <PortfolioTransactionRow
              event={event}
              key={idx}
              security={securities.find(
                (sec) => sec.id == event.securityId
              )}
              currency="EUR"
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}

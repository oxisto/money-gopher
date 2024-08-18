import {
  Currency,
  PortfolioEvent,
  PortfolioEventType,
  Security,
} from "@/lib/gen/mgo_pb";
import FormattedDate from "@/components/formatted-date";
import { currency as formatCurrency } from "@/lib/util";
import { PencilSquareIcon } from "@heroicons/react/24/outline";

interface PortfolioTransactionRowProps {
  event: PortfolioEvent;
  security?: Security;
  currency: string;
}

export default function PortfolioTransactionRow({
  event,
  security,
  currency = "EUR",
}: PortfolioTransactionRowProps) {
  const total = new Currency({
    symbol: currency,
    value:
      event.amount *
      ((event.price?.value ?? 0) +
        (event.fees?.value ?? 0) +
        (event.taxes?.value ?? 0)),
  });

  return (
    <tr>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        <FormattedDate date={event.time?.toDate()} />
      </td>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        {PortfolioEventType[event.type]}
      </td>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        {security && (
          <div className="text-gray-900">{security.displayName}</div>
        )}
        <div className="text-gray-400">{event.securityName}</div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500 md:table-cell">
        {Intl.NumberFormat(navigator.language, {
          maximumFractionDigits: 2,
        }).format(event.amount)}
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
        <div className="text-gray-500">{formatCurrency(event.price)}</div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">{formatCurrency(event.fees)}</div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">{formatCurrency(event.taxes)}</div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">{formatCurrency(total)}</div>
      </td>
      <td>
        <a href="/portfolios/{tx.portfolioName}/transactions/{tx.name}">
          <PencilSquareIcon className="h-5 w-5 text-gray-400" />
        </a>
      </td>
    </tr>
  );
}

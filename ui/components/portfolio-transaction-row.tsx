import FormattedCurrency from "@/components/formatted-currency";
import FormattedDate from "@/components/formatted-date";
import { SchemaCurrency, SchemaPortfolioEvent, SchemaSecurity } from "@/lib/api";
import { PencilSquareIcon } from "@heroicons/react/24/outline";
import Link from "next/link";

interface PortfolioTransactionRowProps {
  event: SchemaPortfolioEvent;
  security?: SchemaSecurity;
  currency: string;
}

export default function PortfolioTransactionRow({
  event,
  security,
  currency = "EUR",
}: PortfolioTransactionRowProps) {
  const total = {
    symbol: currency,
    value:
      event.amount *
      ((event.price?.value ?? 0) +
        (event.fees?.value ?? 0) +
        (event.taxes?.value ?? 0)),
  } satisfies SchemaCurrency;

  return (
    <tr>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        <FormattedDate date={event.time} />
      </td>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        TODO
      </td>
      <td className="truncate whitespace-nowrap py-2 pl-4 pr-3 text-sm font-medium sm:pl-0">
        {security && (
          <div className="text-gray-900">{security.displayName}</div>
        )}
        <div className="text-gray-400">{event.securityId}</div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm text-gray-500 md:table-cell">
        {Intl.NumberFormat(navigator.language, {
          maximumFractionDigits: 2,
        }).format(event.amount)}
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm lg:table-cell">
        <div className="text-gray-500">
          <FormattedCurrency value={event.price} />
        </div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">
          <FormattedCurrency value={event.fees} />
        </div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">
          <FormattedCurrency value={event.taxes} />
        </div>
      </td>
      <td className="hidden whitespace-nowrap px-3 py-2 text-right text-sm sm:table-cell">
        <div className="text-gray-500">
          <FormattedCurrency value={total} />
        </div>
      </td>
      <td>
        <Link
          href={`/portfolios/${event.portfolioId}/transactions/${event.id}`}
        >
          <PencilSquareIcon className="h-5 w-5 text-gray-400" />
        </Link>
      </td>
    </tr>
  );
}

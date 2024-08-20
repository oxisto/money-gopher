import FormattedDate from "@/components/formatted-date";
import PortfolioPerformance from "@/components/portfolio-performance";
import client, { Portfolio } from "@/lib/api";
import {
  ArrowTrendingUpIcon,
  CalendarDaysIcon,
  CreditCardIcon,
  UserCircleIcon,
} from "@heroicons/react/24/outline";
import Link from "next/link";
import FormattedCurrency from "./formatted-currency";

interface PortfolioCardProps {
  /**
   * The portfolio to show.
   */
  portfolio: Portfolio;
}

export default async function PortfolioCard({ portfolio }: PortfolioCardProps) {
  const { data: snapshot } = await client.GET(
    "/v1/portfolios/{portfolioName}/snapshot",
    {
      params: {
        path: {
          portfolioName: portfolio.name,
        },
      },
    },
  );

  if (snapshot) {
    return (
      <div className="lg:col-start-3 lg:row-end-1">
        <h2 className="sr-only">Summary</h2>
        <div className="rounded-lg bg-gray-50 shadow-sm ring-1 ring-gray-900/5">
          <dl className="flex flex-wrap">
            <div className="flex-auto pl-6 pt-6">
              <dt className="text-sm font-semibold leading-6 text-gray-900">
                <Link href={`/portfolios/${portfolio.name}`}>
                  {portfolio.displayName}
                </Link>
              </dt>
              <dd className="mt-1 text-base font-semibold leading-6 text-gray-900">
                {snapshot && snapshot.totalMarketValue && (
                  <FormattedCurrency value={snapshot.totalMarketValue} />
                )}
              </dd>
            </div>
            <div className="flex-none self-end px-6 pt-4">
              <dd className="mx-1 inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20">
                Stocks
              </dd>
              <dd className="mx-1 inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20">
                ETFs
              </dd>
            </div>
            <div className="mt-6 flex w-full flex-none gap-x-4 border-t border-gray-900/5 px-6 pt-6">
              <dt className="flex-none">
                <span className="sr-only">Performance</span>
                <ArrowTrendingUpIcon
                  className="h-6 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </dt>
              <dd className="text-sm font-medium leading-6 text-gray-900">
                <PortfolioPerformance snapshot={snapshot} showIcon={false} />
              </dd>
            </div>
            <div className="mt-4 flex w-full flex-none gap-x-4 px-6">
              <dt className="flex-none">
                <span className="sr-only">Client</span>
                <UserCircleIcon
                  className="h-6 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </dt>
              <dd className="text-sm font-medium leading-6 text-gray-900">
                Current User
              </dd>
            </div>
            <div className="mt-4 flex w-full flex-none gap-x-4 px-6">
              <dt className="flex-none">
                <span className="sr-only">Due date</span>
                <CalendarDaysIcon
                  className="h-6 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </dt>
              <dd className="text-sm leading-6 text-gray-500">
                <FormattedDate date={snapshot?.firstTransactionTime} />
              </dd>
            </div>
            <div className="mt-4 flex w-full flex-none gap-x-4 px-6">
              <dt className="flex-none">
                <span className="sr-only">Status</span>
                <CreditCardIcon
                  className="h-6 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </dt>
              <dd className="text-sm leading-6 text-gray-500">
                Associated with <i>My Bank Account</i>
              </dd>
            </div>
          </dl>
          <div className="mt-6 border-t border-gray-900/5 px-6 py-6">
            <a
              href={"/portfolios/" + portfolio.name}
              className="text-sm font-semibold leading-6 text-gray-900"
            >
              Show positions
              <span aria-hidden="true">&rarr;</span>
            </a>
          </div>
        </div>
      </div>
    );
  }
}

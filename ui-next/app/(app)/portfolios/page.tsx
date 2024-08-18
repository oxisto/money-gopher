import { unstable_noStore as noStore } from "next/cache";
import { portfolioClient } from "@/lib/clients";
import PortfolioCard from "@/components/portfolio-card";
import Link from "next/link";

export default async function Portfolios() {
  noStore();
  const portfolios = await portfolioClient
    .listPortfolios({})
    .then((res) => res.portfolios);

  return (
    <>
      <div className="my-4 border-b border-gray-200 bg-white px-4 py-5 sm:px-6">
        <div className="-ml-4 -mt-4 flex flex-wrap items-center justify-between sm:flex-nowrap">
          <div className="ml-4 mt-4">
            <h3 className="text-base font-semibold leading-6 text-gray-900">
              Portfolios
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              Portfolios are a group of assets.
            </p>
          </div>
          <div className="ml-4 mt-4 flex-shrink-0">
            <Link href="/portfolios/new">
              <button
                type="button"
                className="relative inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
              >
                Create new
              </button>
            </Link>
          </div>
        </div>
      </div>
      <ul
        role="list"
        className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3"
      >
        {portfolios.map((portfolio, idx) => (
          <li className="col-span-1 divide-y divide-gray-200 rounded-lg bg-white shadow" key={idx}>
            <PortfolioCard portfolio={portfolio}></PortfolioCard>
          </li>
        ))}
      </ul>
    </>
  );
}

import "server-only"
import { SidebarItem, SidebarItemData } from "@/components/sidebaritem";
import { auth } from "@/lib/auth";
import { classNames } from "@/lib/util";
import { portfolioClient } from "../lib/client";

const navigation: SidebarItemData[] = [
  { name: "Dashboard", href: "/dashboard" /*, icon: HomeIcon*/ },
  { name: "Securities", href: "/securities" /*, icon: BanknotesIcon */ },
  {
    name: "Portfolios",
    href: "/portfolios" /* icon: FolderIcon, */,
  },
  { name: "Dividends", href: "/dividends" /*, icon: CalendarIcon */ },
  { name: "Performance", href: "/performance" /*, icon: ChartPieIcon */ },
];

const teams = [
  { id: 1, name: "Personal", href: "#", initial: "H", current: false },
  { id: 2, name: "Parents", href: "#", initial: "T", current: false },
  { id: 3, name: "Child", href: "#", initial: "W", current: false },
];

const client = portfolioClient(fetch);

export default async function Sidebar({ isDesktop = false }) {
  const portfolios = await client
    .listPortfolios({})
    .then((res) => res.portfolios);
  navigation[2].children = portfolios.map((p) => {
    return { name: p.displayName, href: `/portfolio/${p.name}` };
  });

  return (
    <div
      className={classNames(
        "flex grow flex-col gap-y-5 overflow-y-auto bg-gray-900",
        isDesktop ? "px-6" : "px-6 pb-2 ring-1 ring-white/10"
      )}
    >
      <div className="flex h-16 shrink-0 items-center">
        <img
          alt="Your Company"
          src="https://tailwindui.com/img/logos/mark.svg?color=indigo&shade=500"
          className="h-8 w-auto"
        />
      </div>
      <nav className="flex flex-1 flex-col">
        <ul role="list" className="flex flex-1 flex-col gap-y-7">
          <li>
            <ul role="list" className="-mx-2 space-y-1">
              {navigation.map((item) => (
                <SidebarItem key={item.name} item={item} />
              ))}
            </ul>
          </li>
          <li>
            <div className="text-xs font-semibold leading-6 text-gray-400">
              Your portfolio groups
            </div>
            <ul role="list" className="-mx-2 mt-2 space-y-1">
              {teams.map((team) => (
                <li key={team.name}>
                  <a
                    href={team.href}
                    className={classNames(
                      team.current
                        ? "bg-gray-800 text-white"
                        : "text-gray-400 hover:bg-gray-800 hover:text-white",
                      "group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6"
                    )}
                  >
                    <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-lg border border-gray-700 bg-gray-800 text-[0.625rem] font-medium text-gray-400 group-hover:text-white">
                      {team.initial}
                    </span>
                    <span className="truncate">{team.name}</span>
                  </a>
                </li>
              ))}
            </ul>
          </li>
          {isDesktop && <SidebarProfile />}
        </ul>
      </nav>
    </div>
  );
}

export async function SidebarProfile() {
  const session = await auth();
  if (!session) return null;

  return (
    <li className="-mx-6 mt-auto">
      <a
        href="#"
        className="flex items-center gap-x-4 px-6 py-3 text-sm font-semibold leading-6 text-white hover:bg-gray-800"
      >
        <img
          alt=""
          src="https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80"
          className="h-8 w-8 rounded-full bg-gray-800"
        />
        <span className="sr-only">Your profile</span>
        <span aria-hidden="true">{session.user?.name}</span>
      </a>
    </li>
  );
}

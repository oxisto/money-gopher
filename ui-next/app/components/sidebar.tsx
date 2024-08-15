import {
  BanknotesIcon,
  CalendarIcon,
  ChartPieIcon,
  FolderIcon,
  HomeIcon,
} from "@heroicons/react/24/outline";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { classNames } from "@/lib/util";

interface SidebarItemData {
  name: string;
  href: string;
  icon?: React.ForwardRefExoticComponent<
    React.PropsWithoutRef<React.SVGProps<SVGSVGElement>> & {
      title?: string;
      titleId?: string;
    } & React.RefAttributes<SVGSVGElement>
  >;
  children?: SidebarItemData[];
  disabled?: boolean;
}

const navigation: SidebarItemData[] = [
  { name: "Dashboard", href: "/dashboard", icon: HomeIcon },
  { name: "Securities", href: "/securities", icon: BanknotesIcon },
  {
    name: "Portfolios",
    href: "/portfolios",
    icon: FolderIcon,
    children: [
      {
        name: "My Portfolio",
        href: "/portfolios/my",
      },
    ],
  },
  { name: "Dividends", href: "/dividends", icon: CalendarIcon },
  { name: "Performance", href: "/performance", icon: ChartPieIcon },
];

const teams = [
  { id: 1, name: "Personal", href: "#", initial: "H", current: false },
  { id: 2, name: "Parents", href: "#", initial: "T", current: false },
  { id: 3, name: "Child", href: "#", initial: "W", current: false },
];

export default function Sidebar({ isDesktop = false }) {
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
          {isDesktop && <SidebarProfile name={"Tom Cook"} />}
        </ul>
      </nav>
    </div>
  );
}

interface SidebarItemProps {
  item: SidebarItemData;

  isSubItem?: boolean;
}

/**
 * Renders an individual sidebar item.
 */
export function SidebarItem({ item, isSubItem }: SidebarItemProps) {
  const pathname = usePathname();
  let current = pathname.startsWith(item.href);

  if (!isSubItem) {
    return (
      <li key={item.name}>
        <Link
          href={item.href}
          className={classNames(
            current
              ? "bg-gray-800 text-white"
              : "text-gray-400 hover:bg-gray-800 hover:text-white",
            "group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6"
          )}
        >
          {item.icon && (
            <>
              <item.icon aria-hidden="true" className="h-6 w-6 shrink-0" />
              {item.name}
            </>
          )}
        </Link>
        <ul className="mt-1">
          {item.children?.map((subItem) => (
            <SidebarItem key={subItem.name} item={subItem} isSubItem={true} />
          ))}
        </ul>
      </li>
    );
  } else {
    return (
      <li key={item.name}>
        <Link
          href={item.href}
          className={classNames(
            current
              ? "bg-gray-700 text-white"
              : "text-gray-300 hover:bg-gray-700 hover:text-white",
            "block rounded-md py-2 pl-9 pr-2 text-sm leading-6"
          )}
        >
          {item.name}
        </Link>
      </li>
    );
  }
}

export interface SidebarProfileProps {
  /**
   * The name of the profile.
   */
  name: string;
}

export function SidebarProfile({ name }: SidebarProfileProps) {
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
        <span aria-hidden="true">{name}</span>
      </a>
    </li>
  );
}

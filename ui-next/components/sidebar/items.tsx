"use client";

import { classNames } from "@/lib/util";
import { BanknotesIcon, CalendarIcon, ChartPieIcon, FolderIcon, HomeIcon } from "@heroicons/react/24/outline";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ForwardRefExoticComponent, SVGProps, RefAttributes } from "react";

type IconType = ForwardRefExoticComponent<Omit<SVGProps<SVGSVGElement>, "ref"> & {
  title?: string;
  titleId?: string;
} & RefAttributes<SVGSVGElement>>

let icons = new Map<string, IconType>([
  ["home", HomeIcon],
  ["banknotes", BanknotesIcon],
  ["folder", FolderIcon],
  ["chartpie", ChartPieIcon],
  ["calendar", CalendarIcon],
]);

export interface SidebarItemData {
  name: string;
  href: string;
  icon?: string;
  children?: SidebarItemData[];
  disabled?: boolean;
}

interface SidebarItemProps {
  item: SidebarItemData;

  isSubItem?: boolean;
}

export function DashboardItem() {
  return <SidebarItem item={
    {name: "Dashboard", href: "/dashboard", icon: "home"  }
  } />
}

export function SecuritiesItem() {
  return <SidebarItem item={
    {name: "Securities", href: "/securities", icon: "banknotes"  }
  } />
}

export function DividendsItem() {
  return <SidebarItem item={
    {name: "Dividends", href: "/dividends", icon: "calendar"  }
  } />
}

export function PerformanceItem() {
  return <SidebarItem item={
    {name: "Performance", href: "/performance", icon: "chartpie"  }
  } />
}

/**
 * Renders an individual sidebar menu item. This must be done on the client,
 * since we need access to the navigation data.
 */
export function SidebarItem({ item, isSubItem }: SidebarItemProps) {
  const pathname = usePathname();
  let current = pathname.startsWith(item.href);
  const Icon = icons.get(item.icon ?? "")

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
          {Icon && (
            <>
              <Icon aria-hidden="true" className="h-6 w-6 shrink-0" />
            </>
          )}
          {item.name}
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
          {Icon && (
            <>
              <Icon aria-hidden="true" className="h-6 w-6 shrink-0" />
            </>
          )}
          {item.name}
        </Link>
      </li>
    );
  }
}

"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { classNames } from "@/lib/util";

export interface SidebarItemData {
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

export interface SidebarItemProps {
  item: SidebarItemData;

  isSubItem?: boolean;
}

/**
 * Renders an individual sidebar menu item. This must be done on the client,
 * since we need access to the navigation data.
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
          {item.icon && (
            <>
              <item.icon aria-hidden="true" className="h-6 w-6 shrink-0" />
            </>
          )}
          {item.name}
        </Link>
      </li>
    );
  }
}

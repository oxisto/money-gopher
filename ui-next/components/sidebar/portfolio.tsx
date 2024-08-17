import { SidebarItem } from "@/components/sidebar/items";
import { portfolioClient } from "@/lib/clients";
import { unstable_noStore as noStore } from "next/cache";

export async function PortfolioItem() {
  noStore();
  const portfolios = await portfolioClient
    .listPortfolios({})
    .then((res) => res.portfolios);

  return (
    <SidebarItem
      item={{
        name: "Portfolios",
        href: "/portfolios",
        icon: "folder",
        children: portfolios.map((p) => {
          return { name: p.displayName, href: `/portfolios/${p.name}` };
        }),
      }}
    />
  );
}

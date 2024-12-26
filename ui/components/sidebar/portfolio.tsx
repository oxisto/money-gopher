import { SidebarItem } from "@/components/sidebar/items";
import client from "@/lib/api";

export async function PortfolioItem() {
  const { data } = await client.GET("/v1/portfolios");
  const portfolios = data?.portfolios ?? [];

  if (data != undefined) {
    return (
      <SidebarItem
        item={{
          name: "Portfolios",
          href: "/portfolios",
          icon: "folder",
          children: portfolios.map((p) => {
            return { name: p.displayName, href: `/portfolios/${p.id}` };
          }),
        }}
      />
    );
  }
}

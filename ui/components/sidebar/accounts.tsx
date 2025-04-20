import { SidebarItem } from "@/components/sidebar/items";
import client from "@/lib/api";

export async function AccountsItem() {
  //const { data } = await client.GET("/v1/accounts");
  //const portfolios = data?.portfolios ?? [];
  const data = {};

  if (data != undefined) {
    return (
      <SidebarItem
        item={{
          name: "Accounts",
          href: "/accounts",
          icon: "folder",
          /*children: portfolios.map((p) => {
            return { name: p.displayName, href: `/portfolios/${p.id}` };
          }),*/
        }}
      />
    );
  }
}

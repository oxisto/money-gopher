import { unstable_noStore as noStore } from 'next/cache';
import { securitiesClient } from "@/lib/clients";

export default async function Securities() {
    noStore();
    const securities = await securitiesClient
      .listSecurities({})
      .then((res) => res.securities);

    return (
      <>
        Securities
        {JSON.stringify(securities)}
      </>
    );
}

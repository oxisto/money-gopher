import client from "@/lib/api";

export default async function Securities() {
  const { data } = await client.GET("/v1/securities");

  return (
    <>
      Securities
      {JSON.stringify(data?.securities)}
    </>
  );
}

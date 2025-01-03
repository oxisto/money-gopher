import client, { SchemaPortfolio } from "@/lib/api";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

export async function createPortfolio(formData: FormData) {
  "use server";

  const portfolio: SchemaPortfolio = {
    id: formData.get("id")?.toString() ?? "",
    displayName: formData.get("displayName")?.toString() ?? "",
    bankAccountId: "",
  };

  const { data: newPortfolio, error } = await client.POST("/v1/portfolios", {
    body: portfolio,
  });
  if (error != undefined) {
    throw error;
  }

  if (newPortfolio) {
    revalidatePath("/portfolios");
    redirect(`/portfolios/${newPortfolio.id}`);
  }
}

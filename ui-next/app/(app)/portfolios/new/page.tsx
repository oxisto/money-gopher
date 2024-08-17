import { portfolioClient } from "@/lib/clients";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

export default function NewPortfolio() {
  async function createPortfolio(formData: FormData) {
    "use server";

    const req = {
      portfolio: {
        name: formData.get("name")?.toString(),
        displayName: formData.get("displayName")?.toString(),
      }
    };

    const portfolio = await portfolioClient.createPortfolio(req)
    if(portfolio) {
      revalidatePath('/portfolios')
      redirect(`/portfolios/${portfolio.name}`)
    }
  }
  return (
    <form action={createPortfolio}>
      <input type="text" name="name" placeholder="Enter name" />
      <input type="text" name="displayName" placeholder="Enter display name" />
      <br />
      <input type="submit" />
    </form>
  );
}

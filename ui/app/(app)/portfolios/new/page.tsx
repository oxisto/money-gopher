import { createPortfolio } from "@/actions/create-portfolio";

export default function NewPortfolio() {
  return (
    <form action={createPortfolio}>
      <input type="text" name="id" placeholder="Enter ID" />
      <input type="text" name="displayName" placeholder="Enter display name" />
      <br />
      <input type="submit" />
    </form>
  );
}

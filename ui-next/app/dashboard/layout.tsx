import Nav from "@/app/ui/nav"

export default function DashboardLayout({
  children, // will be a page or nested layout
}: {
  children: React.ReactNode
}) {
  return (
    <section>
      <Nav />

      <main className="py-10 lg:pl-72">
        <div className="px-4 sm:px-6 lg:px-8">
          {children}
        </div>
      </main>

    </section>
  )
}
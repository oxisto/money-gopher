import Button from "@/components/button";
import { ChevronRightIcon } from "@heroicons/react/20/solid";

export default async function Page() {
  const people = [
    {
      name: "Leslie Alexander",
      email: "leslie.alexander@example.com",
      role: "Co-Founder / CEO",
      imageUrl:
        "https://images.unsplash.com/photo-1494790108377-be9c29b29330?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
      href: "#",
      lastSeen: "3h ago",
      lastSeenDateTime: "2023-01-23T13:23Z",
    },
  ];

  const result = await execute(ListAccountsDocument);

  return (
    <ul
      role="list"
      className="divide-y divide-gray-100 overflow-hidden bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl"
    >
      {JSON.stringify(result)}
      {result.accounts?.map((account) => (
        <li
          key={account.id}
          className="relative flex justify-between gap-x-6 px-4 py-5 hover:bg-gray-50 sm:px-6"
        >
          <div className="flex min-w-0 gap-x-4">
            <img
              alt=""
              src={account.imageUrl}
              className="size-12 flex-none rounded-full bg-gray-50"
            />
            <div className="min-w-0 flex-auto">
              <p className="text-sm/6 font-semibold text-gray-900">
                <a href={account.href}>
                  <span className="absolute inset-x-0 -top-px bottom-0" />
                  {account.displayName}
                </a>
              </p>
              <p className="mt-1 flex text-xs/5 text-gray-500">
                <a
                  href={`mailto:${account.email}`}
                  className="relative truncate hover:underline"
                >
                  {account.email}
                </a>
              </p>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-x-4">
            <div className="hidden sm:flex sm:flex-col sm:items-end">
              <p className="text-sm/6 text-gray-900">{account.role}</p>
              {account.lastSeen ? (
                <p className="mt-1 text-xs/5 text-gray-500">
                  Last seen{" "}
                  <time dateTime={account.lastSeenDateTime}>
                    {account.lastSeen}
                  </time>
                </p>
              ) : (
                <div className="mt-1 flex items-center gap-x-1.5">
                  <div className="flex-none rounded-full bg-emerald-500/20 p-1">
                    <div className="size-1.5 rounded-full bg-emerald-500" />
                  </div>
                  <p className="text-xs/5 text-gray-500">Online</p>
                </div>
              )}
            </div>
            <ChevronRightIcon
              aria-hidden="true"
              className="size-5 flex-none text-gray-400"
            />
          </div>
        </li>
      ))}
      <li>
        <Button>New Account</Button>
      </li>
    </ul>
  );
}

import {
  ListAccountsDocument,
  type TypedDocumentString,
} from "@/lib/gql/graphql";

export async function execute<TResult, TVariables>(
  query: TypedDocumentString<TResult, TVariables>,
  ...[variables]: TVariables extends Record<string, never> ? [] : [TVariables]
) {
  const response = await fetch("http://localhost:8080/graphql/query", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/graphql-response+json",
    },
    body: JSON.stringify({
      query,
      variables,
    }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  //return (response.json() as any).data as TResult;
  return response.json().then((data) => data.data as TResult);
}

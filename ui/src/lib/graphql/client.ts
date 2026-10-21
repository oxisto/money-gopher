import { initGraphQLTada, type TadaDocumentNode } from 'gql.tada';
import { print } from 'graphql';
import type { introspection } from './graphql-env.d.ts';

/**
 * Typed GraphQL documents: `graphql(\`query … \`)` infers the result and
 * variable types from api/schema.graphql — no code generation besides
 * graphql-env.d.ts, which `npm run generate` refreshes from the schema.
 */
export const graphql = initGraphQLTada<{
	introspection: introspection;
	scalars: {
		// Revived from its RFC 3339 string in send() below.
		Time: Date;
		// Kept as the exact decimal string, never parsed to a float.
		Decimal: string;
	};
}>();

export type { FragmentOf, ResultOf, VariablesOf } from 'gql.tada';

// The Time scalar is the only value this API sends as a full RFC 3339
// timestamp; turning exactly those strings into Dates while parsing gives
// every Time field its declared type without a per-query field map.
const RFC3339 = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/;

function revive(_key: string, value: unknown) {
	return typeof value === 'string' && RFC3339.test(value) ? new Date(value) : value;
}

interface Response<Result> {
	data?: Result | null;
	errors?: { message: string }[];
}

async function send<Result, Variables>(
	document: TadaDocumentNode<Result, Variables>,
	variables: Variables,
	fetchFn: typeof fetch
): Promise<Response<Result>> {
	const res = await fetchFn('/graphql', {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({ query: print(document), variables })
	});

	// No valid session: sign in again. /login itself loads no data, so this
	// can't loop.
	if (res.status === 401 && window.location.pathname !== '/login') {
		window.location.href = '/login';
		return new Promise(() => {});
	}

	const text = await res.text();
	try {
		return JSON.parse(text, revive) as Response<Result>;
	} catch {
		throw new Error(`GraphQL request failed (${res.status}): ${text.slice(0, 200)}`);
	}
}

/**
 * Runs a query for a load function. Errors throw, which SvelteKit turns
 * into the error page; pass load's `fetch` so SvelteKit can track it.
 */
export async function query<Result, Variables>(
	document: TadaDocumentNode<Result, Variables>,
	variables: Variables,
	fetchFn: typeof fetch = fetch
): Promise<Result> {
	const { data, errors } = await send(document, variables, fetchFn);
	if (errors?.length) throw new Error(errors[0].message);
	if (!data) throw new Error('GraphQL response without data');
	return data;
}

/**
 * Runs a mutation from a form or button. Errors come back as a message to
 * show next to the control, not as an exception. Afterwards, call
 * invalidateAll() to reload whatever the page shows.
 */
export async function mutate<Result, Variables>(
	document: TadaDocumentNode<Result, Variables>,
	variables: Variables
): Promise<{ data: Result | null; error: string | null }> {
	try {
		const { data, errors } = await send(document, variables, fetch);
		return { data: data ?? null, error: errors?.[0]?.message ?? null };
	} catch (e) {
		return { data: null, error: e instanceof Error ? e.message : String(e) };
	}
}

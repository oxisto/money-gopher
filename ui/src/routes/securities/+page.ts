import { query } from '$lib/graphql/client';
import { SecuritiesQuery } from '$lib/graphql/documents';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => query(SecuritiesQuery, {}, fetch);

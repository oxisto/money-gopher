import { query } from '$lib/graphql/client';
import { PeopleQuery } from '$lib/graphql/documents';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => query(PeopleQuery, {}, fetch);

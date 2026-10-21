import { query } from '#lib/graphql/client.js';
import { PeopleQuery } from '#lib/graphql/documents.js';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => query(PeopleQuery, {}, fetch);

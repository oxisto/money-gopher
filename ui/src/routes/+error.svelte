<script>
	import { page } from '$app/stores';
	import { redirectLogin } from '$lib/oauth';

	page.subscribe((p) => {
		if (p.status == 401) {
			redirectLogin('/');
		}
	});
</script>

{#if $page.status == 404}
	<p>404 Money not found</p>
{:else}
	There was an error:
	<ul>
		<li>{$page.status}</li>
		<li>{$page.error?.message}</li>
	</ul>
{/if}

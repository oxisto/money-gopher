<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="mx-auto mt-16 max-w-md rounded-xl border border-line p-8 text-center">
	<h1 class="text-base font-semibold">
		{page.status === 404 ? 'Not found' : 'Something went wrong'}
	</h1>
	<p class="mt-1 text-sm text-muted">{page.error?.message}</p>
	<a
		href="/"
		class="mt-5 inline-flex h-8.5 items-center rounded-lg border border-line bg-surface px-3 text-sm font-medium shadow-xs hover:bg-raised"
	>
		Back to the overview
	</a>
</div>

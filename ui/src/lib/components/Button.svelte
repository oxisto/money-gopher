<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';

	let clazz: string | undefined | null = null;

	export { clazz as class };

	interface $$Props extends HTMLButtonAttributes {
		class?: string | undefined | null;
	}
</script>

<button
	type="button"
	class="mt-5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500
  focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600
  {clazz ?? ''}
  "
	on:click
	{...$$restProps}
>
	<slot />
</button>

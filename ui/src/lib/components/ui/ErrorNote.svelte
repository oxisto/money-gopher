<script lang="ts">
	let { message }: { message: string | null } = $props();
</script>

{#if message}
	<p class="mt-2 text-sm text-neg">{message}</p>
{/if}

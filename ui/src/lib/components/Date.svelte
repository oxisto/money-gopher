<script lang="ts">
	export let date: Date | undefined;
</script>

{#if date}
	<time datetime={date.toISOString()}>
		{Intl.DateTimeFormat(navigator.language, {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		}).format(date)}
	</time>
{:else}
	-
{/if}

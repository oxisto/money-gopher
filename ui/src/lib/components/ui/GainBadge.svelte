<script lang="ts">
	import { formatPercent } from '$lib/format';

	let { gains }: { gains: number } = $props();
</script>

<span
	class="rounded px-1.5 py-0.5 text-xs font-medium tabular-nums {gains < 0
		? 'bg-red-50 text-red-700'
		: 'bg-green-50 text-green-700'}"
>
	{formatPercent(gains)}
</span>

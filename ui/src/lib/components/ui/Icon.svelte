<script lang="ts">
	import { icons, type IconName } from '$lib/icons';

	let { name, size = 16, strokeWidth = 2 }: { name: IconName; size?: number; strokeWidth?: number } =
		$props();
</script>

<svg
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width={strokeWidth}
	stroke-linecap="round"
	stroke-linejoin="round"
	aria-hidden="true"
>
	<path d={icons[name]} />
</svg>

<script lang="ts">
	const { active, column } = $props<{
		active: Boolean;
		column: string;
	}>();
</script>

<slot />

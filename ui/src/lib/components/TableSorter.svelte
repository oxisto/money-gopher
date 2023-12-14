<svelte:options runes={true} />

<script lang="ts">
	let { active, column } = $props<{
		/**
		 * Whether this column sorter is active or not.
		 */
		active: Boolean;
		/**
		 * The column we are sorting.
		 */

		column: string;
	}>();
</script>

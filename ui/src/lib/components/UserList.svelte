<script lang="ts">
	import type { People$result } from '$houdini';

	let { users }: { users: People$result['users'] } = $props();
</script>

{#if users.length === 0}
	<p class="text-sm text-gray-500">No users yet.</p>
{:else}
	<ul class="divide-y divide-gray-200 rounded-md border border-gray-200 bg-white">
		{#each users as user (user.id)}
			<li class="px-4 py-3 text-sm font-medium">{user.displayName}</li>
		{/each}
	</ul>
{/if}

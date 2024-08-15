<script lang="ts">
	import { Timestamp } from '@bufbuild/protobuf';
	import dayjs from 'dayjs';

	let {
		format = $bindable('YYYY-MM-DD HH:mm'),
		date = $bindable(new Timestamp())
	}: { format: string; date: Timestamp } = $props();

	let internal = $state(dayjs(date.toDate()).format(format));

	$effect(() => {
		if (internal !== '') {
			date = Timestamp.fromDate(dayjs(internal, format).toDate());
		}
	});
</script>

<input
	type="datetime-local"
	name="type"
	id="type"
	class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300
  placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
	bind:value={internal}
/>

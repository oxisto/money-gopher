<script lang="ts">
	import { Timestamp } from '@bufbuild/protobuf';
	import dayjs from 'dayjs';

	export let format = 'YYYY-MM-DD HH:mm';
	export let date: Timestamp | undefined = new Timestamp();
	export let initial = false;

	let internal: string;

	function input(x: Timestamp | undefined) {
		if (initial) {
			return;
		} else {
			// only do this once because otherwise we encounter a bug where the year 0002 is parsed as 1902
			initial = true;
			if (x !== undefined) {
				internal = dayjs(x.toDate()).format(format);
			} else {
				internal == '';
			}
		}
	}

	function output(x: string) {
		if (x !== '') {
			date = Timestamp.fromDate(dayjs(x, format).toDate());
		} else {
			//date = undefined;
		}
	}

	$: input(date);
	$: output(internal);
</script>

<input
	type="datetime-local"
	name="type"
	id="type"
	class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300
  placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
	bind:value={internal}
/>

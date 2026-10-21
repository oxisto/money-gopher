/**
 * Stroke icon paths on a 24×24 grid, drawn by ui/Icon.svelte. Kept as plain
 * path data so any component can pick one by name without an icon library.
 */
export const icons = {
	plus: 'M12 5v14M5 12h14',
	minus: 'M5 12h14',
	coin: 'M12 3a9 9 0 1 0 0 18a9 9 0 1 0 0-18M12 7v10M9.5 9.5h4a1.5 1.5 0 0 1 0 3h-3a1.5 1.5 0 0 0 0 3h4',
	percent: 'M19 5L5 19M7 4.5a2.5 2.5 0 1 0 0 5a2.5 2.5 0 1 0 0-5M17 14.5a2.5 2.5 0 1 0 0 5a2.5 2.5 0 1 0 0-5',
	arrowIn: 'M12 4v12M6 10l6 6 6-6M5 20h14',
	arrowOut: 'M12 20V8M6 14l6-6 6 6M5 4h14',
	receipt: 'M6 3h12v18l-3-2-3 2-3-2-3 2zM9 8h6M9 12h6',
	refund: 'M4 12a8 8 0 1 0 2.3-5.7M4 4v4h4',
	transferIn: 'M4 12h12M12 6l6 6-6 6M20 4v16',
	transferOut: 'M20 12H8M12 6l-6 6 6 6M4 4v16',
	document: 'M14 3H6v18h12V7zM14 3v4h4M9 13h6M9 17h6',
	check: 'M5 12l5 5L20 7',
	upload: 'M12 15V3M7 8l5-5 5 5M4 15v5h16v-5',
	close: 'M6 6l12 12M18 6L6 18',
	search: 'M11 4a7 7 0 1 0 0 14a7 7 0 1 0 0-14M20 20l-3.5-3.5'
} as const;

export type IconName = keyof typeof icons;

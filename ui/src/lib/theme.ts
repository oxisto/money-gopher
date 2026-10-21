/**
 * Light/dark theme preference. "system" follows prefers-color-scheme;
 * the other two pin it via data-theme on <html>, which app.css keys off.
 * The choice lives in localStorage, so it is per browser, not per user.
 */
export type ThemePreference = 'system' | 'light' | 'dark';

const KEY = 'mg-theme';

export function loadTheme(): ThemePreference {
	try {
		const v = localStorage.getItem(KEY);
		if (v === 'light' || v === 'dark') {
			return v;
		}
	} catch {
		// Storage can be unavailable (private mode, blocked site data).
	}
	return 'system';
}

export function applyTheme(pref: ThemePreference) {
	const root = document.documentElement;
	if (pref === 'system') {
		delete root.dataset.theme;
	} else {
		root.dataset.theme = pref;
	}
	try {
		if (pref === 'system') {
			localStorage.removeItem(KEY);
		} else {
			localStorage.setItem(KEY, pref);
		}
	} catch {
		// See loadTheme.
	}
}

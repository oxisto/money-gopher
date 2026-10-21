<script lang="ts">
	// This page is shown when the server returns 401 (no valid session).
	// In dev mode the backend never returns 401, so this page is only reachable
	// in OIDC mode. Clicking the button starts the authorization-code flow.
</script>

<div class="flex min-h-screen items-center justify-center bg-canvas">
	<div class="w-full max-w-sm rounded-xl border border-line bg-surface p-8 text-center shadow-xs">
		<img src="/gopher.png" alt="Money Gopher" class="mx-auto mb-4 h-16 w-16" />
		<h1 class="mb-2 text-xl font-semibold text-fg">Money Gopher</h1>
		<p class="mb-6 text-sm text-muted">Sign in to manage your investments.</p>
		<a
			href="/auth/login"
			class="inline-block rounded-lg bg-accent px-6 py-2.5 text-sm font-medium text-on-accent hover:bg-accent-hover"
		>
			Sign in
		</a>
	</div>
</div>

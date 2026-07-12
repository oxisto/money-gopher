<script lang="ts">
	// This page is shown when the server returns 401 (no valid session).
	// In dev mode the backend never returns 401, so this page is only reachable
	// in OIDC mode. Clicking the button starts the authorization-code flow.
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-full max-w-sm rounded-lg border border-gray-200 bg-white p-8 text-center shadow-sm">
		<img src="/gopher.png" alt="Money Gopher" class="mx-auto mb-4 h-16 w-16" />
		<h1 class="mb-2 text-xl font-semibold text-gray-900">Money Gopher</h1>
		<p class="mb-6 text-sm text-gray-500">Sign in to manage your investments.</p>
		<a
			href="/auth/login"
			class="inline-block rounded-md bg-gray-900 px-6 py-2.5 text-sm font-medium text-white hover:bg-gray-700"
		>
			Sign in
		</a>
	</div>
</div>

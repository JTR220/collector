<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';

	const authApiUrl = import.meta.env.VITE_AUTH_API_BASE_URL ?? 'http://localhost:8080';

	type Mode = 'login' | 'register';
	let mode: Mode = 'login';

	let email = '';
	let password = '';
	let name = '';
	let loading = false;
	let error = '';
	let success = '';

	onMount(() => {
		if ($isAuthenticated) goto('/');
	});

	async function submit() {
		loading = true;
		error = '';
		success = '';

		try {
			if (mode === 'login') {
				const res = await fetch(`${authApiUrl}/login`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ email, password })
				});
				const data = await res.json();
				if (!res.ok) throw new Error(data.error ?? 'Erreur connexion');
				auth.login(data.token, data.user);
				goto('/');
			} else {
				const res = await fetch(`${authApiUrl}/utilisateur`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ name, email, password })
				});
				const data = await res.json();
				if (!res.ok) throw new Error(data.error ?? 'Erreur inscription');
				success = 'Compte créé. Connectez-vous maintenant.';
				mode = 'login';
				email = '';
				password = '';
				name = '';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Erreur inconnue';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Collector.shop | Connexion</title>
</svelte:head>

<div class="flex min-h-[calc(100vh-57px)] items-center justify-center bg-slate-950 px-4 py-12">
	<div class="w-full max-w-md">
		<div class="rounded-[2rem] border border-slate-800 bg-slate-900 p-8 shadow-2xl">
			<p
				class="mb-2 inline-flex rounded-full border border-orange-700 bg-orange-950 px-3 py-1 text-xs font-semibold uppercase tracking-widest text-orange-400"
			>
				Authentification JWT
			</p>
			<h1 class="text-3xl font-black uppercase tracking-tight text-white">
				{mode === 'login' ? 'Connexion' : 'Inscription'}
			</h1>
			<p class="mt-2 text-sm text-slate-400">
				{mode === 'login'
					? 'Accédez à votre collection et participez aux drops exclusifs.'
					: "Créez un compte pour rejoindre collector.shop."}
			</p>

			{#if error}
				<div class="mt-4 rounded-2xl border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-400">
					{error}
				</div>
			{/if}

			{#if success}
				<div
					class="mt-4 rounded-2xl border border-emerald-800 bg-emerald-950 px-4 py-3 text-sm text-emerald-400"
				>
					{success}
				</div>
			{/if}

<form class="mt-6 space-y-4" onsubmit={(e) => { e.preventDefault(); submit(); }}>
				{#if mode === 'register'}
					<input
						class="w-full rounded-2xl border border-slate-700 bg-slate-800 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:outline-none"
						type="text"
						bind:value={name}
						placeholder="Nom complet"
						required
					/>
				{/if}
				<input
					class="w-full rounded-2xl border border-slate-700 bg-slate-800 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:outline-none"
					type="email"
					bind:value={email}
					placeholder="Email"
					required
				/>
				<input
					class="w-full rounded-2xl border border-slate-700 bg-slate-800 px-4 py-3 text-white placeholder:text-slate-500 focus:border-orange-400 focus:outline-none"
					type="password"
					bind:value={password}
					placeholder="Mot de passe"
					required
				/>
				<button
					type="submit"
					disabled={loading}
					class="w-full rounded-2xl bg-orange-500 px-5 py-3 text-sm font-black uppercase tracking-widest text-white transition hover:bg-orange-400 disabled:cursor-not-allowed disabled:bg-orange-300"
				>
					{loading ? '...' : mode === 'login' ? 'Se connecter' : "S'inscrire"}
				</button>
			</form>

			<div class="mt-6 border-t border-slate-800 pt-4 text-center">
				<button
					onclick={() => {
						mode = mode === 'login' ? 'register' : 'login';
						error = '';
						success = '';
					}}
					class="text-xs text-slate-400 hover:text-white"
				>
					{mode === 'login' ? 'Pas encore de compte ? S\'inscrire' : 'Déjà un compte ? Se connecter'}
				</button>
			</div>

			<div class="mt-6 rounded-2xl border border-slate-700 bg-slate-800 p-4 text-xs text-slate-400">
				<p class="font-bold text-slate-300">Routes protégées (JWT requis)</p>
				<ul class="mt-2 space-y-1">
					<li><span class="text-amber-400">POST</span> /article — créer un article</li>
					<li><span class="text-blue-400">PUT</span> /article/:id — modifier</li>
					<li><span class="text-red-400">DELETE</span> /article/:id — supprimer</li>
					<li><span class="text-amber-400">POST</span> /category — créer une catégorie</li>
					<li><span class="text-emerald-400">GET</span> /me — profil utilisateur</li>
				</ul>
			</div>
		</div>
	</div>
</div>

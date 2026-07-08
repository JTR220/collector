<script lang="ts">
	import { env } from '$env/dynamic/public';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';

	const authApiUrl = env.PUBLIC_AUTH_API_BASE_URL ?? 'http://localhost:8080';

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

	function isValidEmail(value: string) {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
	}

	async function submit() {
		error = '';
		success = '';

		if (!isValidEmail(email)) {
			error = 'Adresse email invalide.';
			return;
		}

		loading = true;

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
				goto(data.user?.role === 'admin' ? '/admin' : '/');
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
	<title>Collector.shop | {mode === 'login' ? 'Connexion' : 'Inscription'}</title>
</svelte:head>

<div class="page">
	<div class="page-left">
		<div class="brand">
			<span class="brand-diamond"></span>
			<span class="brand-name">Collector<span class="brand-dim">.shop</span></span>
		</div>
		<div class="page-left-body">
			<h2 class="page-tagline">La maison des<br />collectionneurs.</h2>
			<p class="page-desc">
				Cartes, consoles, comics, vinyles — évaluez, échangez, et grimpez dans la ligue des
				meilleurs chasseurs.
			</p>
			<div class="page-stats">
				<div class="page-stat">
					<span class="page-stat-val">12 400</span>
					<span class="page-stat-label">pièces référencées</span>
				</div>
				<div class="page-stat">
					<span class="page-stat-val">3 200</span>
					<span class="page-stat-label">collectionneurs actifs</span>
				</div>
				<div class="page-stat">
					<span class="page-stat-val">Saison 03</span>
					<span class="page-stat-label">en cours · ligue active</span>
				</div>
			</div>
		</div>
	</div>

	<div class="page-right">
		<div class="card">
			<div class="card-eyebrow">
				{mode === 'login' ? 'Connexion' : 'Inscription'}
			</div>
			<h1 class="card-title">
				{#if mode === 'login'}Accès<br />Galerie{:else}Nouveau<br />Compte{/if}
			</h1>
			<p class="card-subtitle">
				{mode === 'login'
					? 'Entrez vos identifiants pour rejoindre la ligue.'
					: 'Créez un compte pour participer aux drops exclusifs.'}
			</p>

			{#if error}
				<div class="msg msg-error">{error}</div>
			{/if}
			{#if success}
				<div class="msg msg-success">{success}</div>
			{/if}

			<form
				onsubmit={(e) => {
					e.preventDefault();
					submit();
				}}
			>
				{#if mode === 'register'}
					<div class="field">
						<label class="field-label" for="name">Nom</label>
						<input
							id="name"
							class="field-input"
							type="text"
							bind:value={name}
							placeholder="Votre nom"
							required
						/>
					</div>
				{/if}
				<div class="field">
					<label class="field-label" for="email">Email</label>
					<input
						id="email"
						class="field-input"
						type="email"
						bind:value={email}
						placeholder="collector@example.com"
						required
					/>
				</div>
				<div class="field">
					<label class="field-label" for="password">Mot de passe</label>
					<input
						id="password"
						class="field-input"
						type="password"
						bind:value={password}
						placeholder="••••••••"
						required
					/>
				</div>
				<button type="submit" class="submit-btn" disabled={loading}>
					{#if loading}
						<span class="spinner"></span>
					{:else}
						{mode === 'login' ? 'Se connecter' : 'Créer le compte'}
					{/if}
				</button>
			</form>

			<div class="divider"></div>

			<button
				class="toggle-btn"
				onclick={() => {
					mode = mode === 'login' ? 'register' : 'login';
					error = '';
					success = '';
				}}
			>
				{mode === 'login' ? "Pas de compte ? · S'inscrire" : 'Déjà un compte ? · Se connecter'}
			</button>

			<div class="demo">
				<span class="demo-label">Comptes de démo</span>
				<button
					type="button"
					class="demo-row"
					onclick={() => {
						email = 'admin@collector.shop';
						password = 'admin123';
					}}
				>
					<span class="demo-role">Admin</span>
					<span class="demo-creds">admin@collector.shop · admin123</span>
				</button>
				<button
					type="button"
					class="demo-row"
					onclick={() => {
						email = 'test@collector.shop';
						password = 'test123';
					}}
				>
					<span class="demo-role">Test</span>
					<span class="demo-creds">test@collector.shop · test123</span>
				</button>
				<button
					type="button"
					class="demo-row"
					onclick={() => {
						email = 'vendeur@collector.shop';
						password = 'vendeur123';
					}}
				>
					<span class="demo-role">Vendeur</span>
					<span class="demo-creds">vendeur@collector.shop · vendeur123</span>
				</button>
				<button
					type="button"
					class="demo-row"
					onclick={() => {
						email = 'acheteur@collector.shop';
						password = 'acheteur123';
					}}
				>
					<span class="demo-role">Acheteur</span>
					<span class="demo-creds">acheteur@collector.shop · acheteur123</span>
				</button>
				<button
					type="button"
					class="demo-row"
					onclick={() => {
						email = 'vault@collector.shop';
						password = 'vault123';
					}}
				>
					<span class="demo-role">Collector Vault</span>
					<span class="demo-creds">vault@collector.shop · vault123</span>
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	.page {
		min-height: 100vh;
		display: grid;
		grid-template-columns: 1fr 1fr;
		background: #191714;
	}
	@media (max-width: 768px) {
		.page {
			grid-template-columns: 1fr;
		}
	}

	/* ── Left panel ── */
	.page-left {
		background: #211e1a;
		border-right: 1px solid rgba(236, 229, 218, 0.1);
		display: flex;
		flex-direction: column;
		padding: 40px 48px;
	}
	@media (max-width: 768px) {
		.page-left {
			display: none;
		}
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.brand-diamond {
		width: 9px;
		height: 9px;
		background: #86b3a4;
		border-radius: 2px;
		transform: rotate(45deg);
		display: inline-block;
	}
	.brand-name {
		font-family: 'Newsreader', Georgia, serif;
		font-size: 20px;
		color: #ece5da;
	}
	.brand-dim {
		color: #766d60;
	}

	.page-left-body {
		margin-top: auto;
		padding-bottom: 40px;
	}
	.page-tagline {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 48px;
		line-height: 1.05;
		color: #ece5da;
		margin: 0 0 18px;
	}
	.page-desc {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 15px;
		color: #a39a8c;
		line-height: 1.6;
		max-width: 400px;
		margin: 0 0 32px;
	}
	.page-stats {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.page-stat {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.page-stat-val {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 22px;
		color: #ece5da;
	}
	.page-stat-label {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		color: #766d60;
		letter-spacing: 0.04em;
	}

	/* ── Right panel ── */
	.page-right {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 40px 24px;
	}

	.card {
		width: 100%;
		max-width: 400px;
	}
	.card-eyebrow {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		font-weight: 600;
		color: #86b3a4;
		margin-bottom: 14px;
	}
	.card-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 46px;
		line-height: 0.95;
		color: #ece5da;
		margin: 0 0 12px;
	}
	.card-subtitle {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		color: #a39a8c;
		line-height: 1.5;
		margin-bottom: 24px;
	}

	/* Messages */
	.msg {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		margin-bottom: 16px;
		line-height: 1.4;
	}
	.msg-error {
		border-color: rgba(215, 156, 134, 0.3);
		background: rgba(215, 156, 134, 0.06);
		color: #d79c86;
	}
	.msg-success {
		border-color: rgba(134, 192, 153, 0.3);
		background: rgba(134, 192, 153, 0.06);
		color: #86c099;
	}

	/* Champs */
	.field {
		margin-bottom: 14px;
	}
	.field-label {
		display: block;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
		margin-bottom: 6px;
	}
	.field-input {
		width: 100%;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 7px;
		padding: 11px 14px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		outline: none;
		transition:
			border-color 150ms,
			box-shadow 150ms;
		box-sizing: border-box;
	}
	.field-input::placeholder {
		color: rgba(236, 229, 218, 0.2);
	}
	.field-input:focus {
		border-color: rgba(134, 179, 164, 0.5);
		box-shadow: 0 0 0 3px rgba(134, 179, 164, 0.08);
	}

	/* Bouton */
	.submit-btn {
		width: 100%;
		margin-top: 6px;
		padding: 13px;
		border-radius: 7px;
		border: none;
		background: #86b3a4;
		color: #191714;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition:
			filter 150ms,
			opacity 150ms;
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 46px;
	}
	.submit-btn:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.submit-btn:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}

	.spinner {
		width: 16px;
		height: 16px;
		border: 2px solid rgba(25, 23, 20, 0.3);
		border-top-color: #191714;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.divider {
		border: none;
		border-top: 1px solid rgba(236, 229, 218, 0.08);
		margin: 22px 0 16px;
	}

	.toggle-btn {
		background: none;
		border: none;
		color: #766d60;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12px;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition: color 150ms;
		padding: 0;
		width: 100%;
		text-align: center;
	}
	.toggle-btn:hover {
		color: #86b3a4;
	}

	/* Encart comptes de démo */
	.demo {
		margin-top: 20px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 9px;
		padding: 12px;
		background: rgba(255, 255, 255, 0.02);
	}
	.demo-label {
		display: block;
		font-size: 10.5px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: #766d60;
		margin-bottom: 8px;
	}
	.demo-row {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		background: none;
		border: none;
		border-radius: 6px;
		padding: 7px 8px;
		cursor: pointer;
		text-align: left;
		transition: background 120ms;
	}
	.demo-row:hover {
		background: rgba(134, 179, 164, 0.08);
	}
	.demo-role {
		flex-shrink: 0;
		font-size: 10.5px;
		font-weight: 700;
		letter-spacing: 0.04em;
		color: #86b3a4;
		width: 58px;
	}
	.demo-creds {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #a39a8c;
	}
</style>

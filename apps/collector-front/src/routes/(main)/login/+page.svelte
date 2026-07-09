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
		background: var(--c-bg);
	}
	@media (max-width: 768px) {
		.page {
			grid-template-columns: 1fr;
		}
	}

	/* ── Left panel ── */
	.page-left {
		background: linear-gradient(135deg, #1e3b2c, #2a4e3a);
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
		background: var(--c-accent);
		border-radius: 2px;
		transform: rotate(45deg);
		display: inline-block;
	}
	.brand-name {
		font-family: var(--f-serif);
		font-weight: 700;
		font-size: 20px;
		color: var(--c-bg);
	}
	.brand-dim {
		color: #c9e0ce;
	}

	.page-left-body {
		margin-top: auto;
		padding-bottom: 40px;
	}
	.page-tagline {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 44px;
		line-height: 1.1;
		color: var(--c-bg);
		margin: 0 0 18px;
	}
	.page-desc {
		font-family: var(--f-body);
		font-size: 15px;
		color: #d8e6db;
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
		font-family: var(--f-serif);
		font-size: 22px;
		font-weight: 600;
		color: var(--c-bg);
	}
	.page-stat-label {
		font-family: var(--f-body);
		font-size: 12px;
		color: #c9e0ce;
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
		font-family: var(--f-body);
		font-size: 11px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		font-weight: 600;
		color: var(--c-ink);
		margin-bottom: 14px;
	}
	.card-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 40px;
		line-height: 1.05;
		color: var(--c-text);
		margin: 0 0 12px;
	}
	.card-subtitle {
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		line-height: 1.5;
		margin-bottom: 24px;
	}

	/* Messages */
	.msg {
		padding: 10px 14px;
		border-radius: 7px;
		border: 1px solid;
		font-family: var(--f-body);
		font-size: 13px;
		margin-bottom: 16px;
		line-height: 1.4;
	}
	.msg-error {
		border-color: rgba(176, 67, 42, 0.3);
		background: #fbe9e3;
		color: var(--c-error);
	}
	.msg-success {
		border-color: rgba(63, 122, 82, 0.3);
		background: var(--c-badge-verified-bg);
		color: var(--c-success);
	}

	/* Champs */
	.field {
		margin-bottom: 14px;
	}
	.field-label {
		display: block;
		font-family: var(--f-body);
		font-size: 11px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--c-text-muted);
		margin-bottom: 6px;
	}
	.field-input {
		width: 100%;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 7px;
		padding: 11px 14px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 14px;
		outline: none;
		transition: border-color 150ms;
		box-sizing: border-box;
	}
	.field-input::placeholder {
		color: var(--c-text-muted);
	}
	.field-input:focus {
		border-color: var(--c-ink);
	}

	/* Bouton */
	.submit-btn {
		width: 100%;
		margin-top: 6px;
		padding: 13px;
		border-radius: 7px;
		border: none;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
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
		border: 2px solid rgba(255, 255, 255, 0.35);
		border-top-color: #fff;
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
		border-top: 1px solid var(--c-border);
		margin: 22px 0 16px;
	}

	.toggle-btn {
		background: none;
		border: none;
		color: var(--c-text-muted);
		font-family: var(--f-body);
		font-size: 12px;
		letter-spacing: 0.04em;
		cursor: pointer;
		transition: color 150ms;
		padding: 0;
		width: 100%;
		text-align: center;
	}
	.toggle-btn:hover {
		color: var(--c-ink);
	}

	/* Encart comptes de démo */
	.demo {
		margin-top: 20px;
		border: 1px solid var(--c-border);
		border-radius: 9px;
		padding: 12px;
		background: var(--c-bg);
	}
	.demo-label {
		display: block;
		font-size: 10.5px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--c-text-muted);
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
		background: var(--c-badge-verified-bg);
	}
	.demo-role {
		flex-shrink: 0;
		font-size: 10.5px;
		font-weight: 700;
		letter-spacing: 0.04em;
		color: var(--c-ink);
		width: 58px;
	}
	.demo-creds {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-text-muted);
	}
</style>

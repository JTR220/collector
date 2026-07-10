<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { messages } from '$lib/stores/messages';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';

	let { children } = $props();

	let loading = $state(true);
	let query = $state('');

	const activeId = $derived($page.params.id ?? '');

	const filtered = $derived(
		$messages.conversations.filter((c) => {
			const q = query.trim().toLowerCase();
			if (!q) return true;
			return (
				(c.other_user_name ?? '').toLowerCase().includes(q) ||
				(c.article_name ?? '').toLowerCase().includes(q) ||
				(c.last_message ?? '').toLowerCase().includes(q)
			);
		})
	);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.user) {
			goto('/login');
			return;
		}
		await messages.refresh();
		loading = false;
	});

	const fmtDate = (iso: string) => {
		const d = new Date(iso);
		const now = new Date();
		if (d.toDateString() === now.toDateString()) {
			return d.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
		}
		return d.toLocaleDateString('fr-FR', { day: '2-digit', month: '2-digit' });
	};
</script>

<svelte:head><title>Messages · Collector.shop</title></svelte:head>

{#if loading}
	<p class="state-msg">Chargement…</p>
{:else}
	<GPanel style="padding:0;margin-top:16px;overflow:hidden">
		<div class="messagerie" class:has-active={!!activeId}>
			<div class="list-pane">
				<div class="list-head">
					<h1 class="list-title">Messages</h1>
					<input
						type="search"
						placeholder="Rechercher une conversation…"
						bind:value={query}
						aria-label="Rechercher une conversation"
					/>
				</div>
				<div class="conv-list">
					{#each filtered as c (c.conversation_id)}
						<a
							class="conv-row"
							class:active={c.conversation_id === activeId}
							href={`/messages/${c.conversation_id}`}
						>
							<GAvatar initials={c.other_user_name?.slice(0, 2).toUpperCase() ?? '??'} size={40} />
							<div class="conv-body">
								<div class="conv-top">
									<span class="conv-name">{c.other_user_name || 'Utilisateur'}</span>
									<span class="conv-date">{fmtDate(c.last_at)}</span>
								</div>
								{#if c.article_name}
									<span class="conv-article">à propos de « {c.article_name} »</span>
								{/if}
								<span class="conv-preview">{c.last_message}</span>
							</div>
							{#if c.unread_count > 0}
								<span class="conv-badge">{c.unread_count}</span>
							{/if}
						</a>
					{:else}
						<p class="item-empty">
							Aucune conversation pour l'instant. Contactez un vendeur depuis une fiche produit.
						</p>
					{/each}
				</div>
			</div>

			<div class="thread-pane">
				{@render children()}
			</div>
		</div>
	</GPanel>
{/if}

<style>
	.state-msg {
		text-align: center;
		padding: 60px 0;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 15px;
		color: var(--c-text-muted);
	}
	.messagerie {
		display: flex;
		height: calc(100vh - 220px);
		min-height: 480px;
	}
	.list-pane {
		width: 340px;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		border-right: 1px solid var(--c-border);
		min-width: 0;
	}
	.list-head {
		flex-shrink: 0;
		padding: 18px 18px 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.list-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 22px;
		color: var(--c-text);
		margin: 0;
	}
	.list-head input {
		width: 100%;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: var(--r-pill);
		padding: 9px 14px;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text);
	}
	.list-head input:focus {
		outline: none;
		border-color: var(--c-ink);
	}
	.conv-list {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
	}
	.conv-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 18px;
		border-top: 1px solid var(--c-border);
		text-decoration: none;
		transition: background 120ms;
	}
	.conv-row:hover {
		background: var(--c-bg);
	}
	.conv-row.active {
		background: var(--c-badge-verified-bg);
	}
	.conv-body {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.conv-top {
		display: flex;
		justify-content: space-between;
		gap: 10px;
	}
	.conv-name {
		font-family: var(--f-body);
		font-size: 14px;
		font-weight: 600;
		color: var(--c-text);
	}
	.conv-date {
		font-family: var(--f-body);
		font-size: 11px;
		color: var(--c-text-muted);
		flex-shrink: 0;
	}
	.conv-article {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-ink);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.conv-preview {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.conv-badge {
		min-width: 20px;
		padding: 2px 6px;
		border-radius: var(--r-pill);
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 700;
		text-align: center;
		flex-shrink: 0;
	}
	.item-empty {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-muted);
		line-height: 1.5;
		padding: 16px 18px;
	}
	.thread-pane {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	@media (max-width: 880px) {
		.messagerie {
			height: calc(100vh - 220px);
		}
		.messagerie.has-active .list-pane {
			display: none;
		}
		.messagerie:not(.has-active) .thread-pane {
			display: none;
		}
		.list-pane {
			width: 100%;
			border-right: none;
		}
	}
</style>

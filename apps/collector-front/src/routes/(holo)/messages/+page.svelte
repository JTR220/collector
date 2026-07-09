<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import { messages } from '$lib/stores/messages';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let loading = $state(true);

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		await messages.refresh($auth.token);
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

<Kicker>Messagerie</Kicker>
<h1 class="page-title">Conversations</h1>

{#if loading}
	<p class="state-msg">Chargement…</p>
{:else}
	<GPanel style="margin-top:14px">
		<div class="conv-list">
			{#each $messages.conversations as c (c.conversation_id)}
				<a class="conv-row" href={`/messages/${c.conversation_id}`}>
					<div class="conv-avatar">{c.other_user_name?.slice(0, 2).toUpperCase() ?? '??'}</div>
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
	.page-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 34px;
		color: var(--c-text);
		margin: 8px 0 0;
	}
	.conv-list {
		display: flex;
		flex-direction: column;
	}
	.conv-row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 4px;
		border-bottom: 1px solid var(--c-border);
		text-decoration: none;
	}
	.conv-avatar {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--c-ink);
		color: var(--c-bg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		flex-shrink: 0;
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
		padding: 12px 4px;
	}
</style>

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
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 12px;
		color: #766d60;
		letter-spacing: 0.12em;
	}
	.page-title {
		font-family: 'Newsreader', Georgia, serif;
		font-weight: 500;
		font-size: 38px;
		color: #ece5da;
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
		border-bottom: 1px solid rgba(236, 229, 218, 0.1);
		text-decoration: none;
	}
	.conv-avatar {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 1px solid rgba(236, 229, 218, 0.16);
		background: rgba(255, 255, 255, 0.05);
		color: #86b3a4;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 14px;
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
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 14px;
		font-weight: 600;
		color: #ece5da;
	}
	.conv-date {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		color: #766d60;
		flex-shrink: 0;
	}
	.conv-article {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 11.5px;
		color: #86b3a4;
	}
	.conv-preview {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #a39a8c;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.conv-badge {
		min-width: 20px;
		padding: 2px 6px;
		border-radius: 999px;
		background: #86b3a4;
		color: #191714;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 11px;
		font-weight: 700;
		text-align: center;
		flex-shrink: 0;
	}
	.item-empty {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		color: #766d60;
		line-height: 1.5;
		padding: 12px 4px;
	}
</style>

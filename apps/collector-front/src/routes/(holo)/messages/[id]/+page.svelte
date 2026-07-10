<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, isAuthenticated } from '$lib/stores/auth';
	import {
		fetchConversationMessages,
		sendMessage,
		markConversationRead,
		connectMessages,
		toUserUUID,
		type MessageAPI,
		type MessageSocket
	} from '$lib/api/messages';
	import { fetchArticle, articleImage, type ArticleAPI } from '$lib/api/catalog';
	import { messages as messagesStore } from '$lib/stores/messages';
	import { fromEventUuid } from '$lib/utils/eventId';
	import { eur } from '$lib/utils/format';
	import GAvatar from '$lib/components/galerie/GAvatar.svelte';

	type BlockedAttempt = { id: string; body: string; created_at: string };

	// Détection cote client de partage de coordonnées personnelles (téléphone / email) :
	// evite un aller-retour reseau pour les cas evidents. notification-service applique
	// aussi son propre filtre cote serveur (internal/pii) — rejet 400 meme si ce filtre
	// local laissait passer un cas.
	const EMAIL_RE = /[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}/i;
	const PHONE_RE = /(?:\+?\d[\s.-]?){7,}\d/;
	const containsContactInfo = (text: string) => EMAIL_RE.test(text) || PHONE_RE.test(text);

	let thread = $state<MessageAPI[]>([]);
	let blockedAttempts = $state<BlockedAttempt[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let draft = $state('');
	let sending = $state(false);
	let socket: MessageSocket | null = null;
	let scrollEl: HTMLDivElement | null = null;
	let article = $state<ArticleAPI | null>(null);
	let lastThreadLen = 0;

	const conversationId = $derived($page.params.id ?? '');
	const myUUID = $derived($auth.user ? toUserUUID($auth.user.id) : '');
	const otherName = $derived(
		thread.length > 0
			? thread[0].sender_id === myUUID
				? thread[0].recipient_name
				: thread[0].sender_name
			: '…'
	);
	const otherInitials = $derived(
		otherName && otherName !== '…' ? otherName.slice(0, 2).toUpperCase() : '??'
	);
	const articleRef = $derived(thread.find((m) => m.article_id));
	const articleName = $derived(articleRef?.article_name ?? null);
	const articleId = $derived(articleRef?.article_id ? fromEventUuid(articleRef.article_id) : null);

	$effect(() => {
		if (thread.length !== lastThreadLen) {
			lastThreadLen = thread.length;
			blockedAttempts = [];
		}
	});

	$effect(() => {
		if (!articleId) {
			article = null;
			return;
		}
		fetchArticle(articleId)
			.then((a) => (article = a))
			.catch(() => (article = null));
	});

	async function scrollToBottom() {
		await tick();
		scrollEl?.scrollTo({ top: scrollEl.scrollHeight });
	}

	onMount(async () => {
		if (!$isAuthenticated || !$auth.user) {
			goto('/login');
			return;
		}
		try {
			thread = await fetchConversationMessages(conversationId);
			await markConversationRead(conversationId);
			messagesStore.refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger la conversation.';
		} finally {
			loading = false;
			scrollToBottom();
		}

		socket = connectMessages((msg) => {
			if (msg.conversation_id !== conversationId) return;
			thread = [...thread, msg];
			scrollToBottom();
			if (msg.recipient_id === myUUID) {
				markConversationRead(conversationId);
				messagesStore.refresh();
			}
		});
	});

	onDestroy(() => socket?.close());

	async function send() {
		const body = draft.trim();
		if (!$auth.user || !body || thread.length === 0) return;

		if (containsContactInfo(body)) {
			blockedAttempts = [
				...blockedAttempts,
				{ id: `blocked-${Date.now()}`, body, created_at: new Date().toISOString() }
			];
			draft = '';
			scrollToBottom();
			return;
		}

		const first = thread[0];
		const recipientId = first.sender_id === myUUID ? first.recipient_id : first.sender_id;

		sending = true;
		try {
			const sent = await sendMessage({
				recipientId,
				body,
				articleId: first.article_id,
				articleName: first.article_name
			});
			thread = [...thread, sent];
			draft = '';
			scrollToBottom();
			messagesStore.refresh();
		} catch (e) {
			error = e instanceof Error ? e.message : "Erreur lors de l'envoi.";
		} finally {
			sending = false;
		}
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			send();
		}
	}

	const fmtTime = (iso: string) =>
		new Date(iso).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
</script>

<svelte:head><title>{otherName} · Messages · Collector.shop</title></svelte:head>

{#if loading}
	<div class="state-wrap"><p class="state-msg">Chargement…</p></div>
{:else if error && thread.length === 0}
	<div class="state-wrap"><p class="state-msg error">{error}</p></div>
{:else}
	<div class="thread-head">
		<a class="back-link" href="/messages">← Conversations</a>
		<div class="thread-head-row">
			<GAvatar initials={otherInitials} size={40} />
			<div class="thread-id">
				<span class="thread-name">{otherName}</span>
				<span class="thread-sub">Particulier vérifié · répond en général sous 1h</span>
			</div>
			<span class="thread-verified-chip">✓ Vérifié par Collector</span>
		</div>
	</div>

	{#if articleName}
		<a class="article-strip" href={articleId ? `/lot/${articleId}` : '#'}>
			<div class="article-thumb">
				{#if article && articleImage(article)}
					<img src={articleImage(article)} alt="" />
				{:else}
					<span class="article-thumb-fallback">◆</span>
				{/if}
			</div>
			<div class="article-info">
				<span class="article-name">{articleName}</span>
				{#if article}<span class="article-price">{eur(article.prix)}</span>{/if}
			</div>
			<span class="article-link">Voir l'annonce</span>
		</a>
	{/if}

	<div class="thread-scroll" bind:this={scrollEl}>
		{#each thread as m (m.id)}
			<div class="msg-row" class:mine={m.sender_id === myUUID}>
				<div class="msg-bubble">
					<span class="msg-text">{m.body}</span>
					<span class="msg-time">{fmtTime(m.created_at)}</span>
				</div>
			</div>
		{/each}
		{#each blockedAttempts as b (b.id)}
			<div class="msg-row mine">
				<div class="msg-bubble blocked">
					<span class="msg-text strike">{b.body}</span>
					<span class="blocked-warning-line">
						<span class="blocked-ico">⊘</span> Message bloqué — le partage de coordonnées personnelles
						n'est pas autorisé sur Collector.shop
					</span>
				</div>
			</div>
			<div class="blocked-info">
				Collector.shop filtre automatiquement les échanges de coordonnées personnelles pour votre
				sécurité.
			</div>
		{/each}
	</div>

	<div class="composer">
		<textarea
			placeholder="Écrire un message…"
			bind:value={draft}
			onkeydown={onKeydown}
			disabled={sending}
			rows="1"></textarea>
		<button class="send-btn" disabled={sending || !draft.trim()} onclick={send}>Envoyer</button>
	</div>
	{#if error}<p class="error-msg">{error}</p>{/if}
{/if}

<style>
	.state-wrap {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.state-msg {
		text-align: center;
		font-family: var(--f-serif);
		font-style: italic;
		font-size: 15px;
		color: var(--c-text-muted);
	}
	.state-msg.error {
		color: var(--c-error);
	}
	.back-link {
		display: none;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		text-decoration: none;
		margin-bottom: 10px;
	}
	.back-link:hover {
		color: var(--c-ink);
	}
	.thread-head {
		flex-shrink: 0;
		padding: 16px 20px;
		border-bottom: 1px solid var(--c-border);
	}
	.thread-head-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.thread-id {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
	}
	.thread-name {
		font-family: var(--f-serif);
		font-size: 17px;
		font-weight: 600;
		color: var(--c-text);
	}
	.thread-sub {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
	}
	.thread-verified-chip {
		flex-shrink: 0;
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		color: var(--c-ink);
		background: var(--c-badge-verified-bg);
		border-radius: var(--r-pill);
		padding: 5px 12px;
		white-space: nowrap;
	}
	.article-strip {
		flex-shrink: 0;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 20px;
		background: var(--c-badge-moderation-bg);
		border-bottom: 1px solid var(--c-border);
		text-decoration: none;
	}
	.article-thumb {
		width: 40px;
		height: 40px;
		flex-shrink: 0;
		border-radius: 8px;
		overflow: hidden;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.article-thumb img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.article-thumb-fallback {
		color: var(--c-icon-muted);
		font-size: 16px;
	}
	.article-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
	}
	.article-name {
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		color: var(--c-text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.article-price {
		font-family: var(--f-serif);
		font-size: 13px;
		color: var(--c-ink);
	}
	.article-link {
		flex-shrink: 0;
		font-family: var(--f-body);
		font-size: 12px;
		font-weight: 600;
		color: var(--c-accent);
	}
	.article-strip:hover .article-link {
		text-decoration: underline;
	}
	.thread-scroll {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px 20px;
	}
	.msg-row {
		display: flex;
		justify-content: flex-start;
	}
	.msg-row.mine {
		justify-content: flex-end;
	}
	.msg-bubble {
		max-width: 70%;
		display: flex;
		flex-direction: column;
		gap: 3px;
		padding: 9px 12px;
		border-radius: 12px;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
	}
	.msg-row.mine .msg-bubble {
		background: var(--c-badge-verified-bg);
		border-color: #cfe3d3;
	}
	.msg-row.mine .msg-bubble.blocked {
		max-width: 78%;
		background: #fbe9e3;
		border-color: rgba(176, 67, 42, 0.35);
	}
	.msg-text {
		font-family: var(--f-body);
		font-size: 13.5px;
		color: var(--c-text);
		white-space: pre-wrap;
		word-break: break-word;
	}
	.msg-text.strike {
		text-decoration: line-through;
		color: var(--c-text-muted);
	}
	.msg-time {
		font-family: var(--f-body);
		font-size: 9.5px;
		color: var(--c-text-muted);
		align-self: flex-end;
	}
	.blocked-warning-line {
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-error);
		line-height: 1.4;
		padding-top: 4px;
		border-top: 1px solid rgba(176, 67, 42, 0.25);
	}
	.blocked-ico {
		font-weight: 700;
	}
	.blocked-info {
		align-self: center;
		max-width: 80%;
		text-align: center;
		font-family: var(--f-body);
		font-size: 11.5px;
		color: var(--c-text-muted);
		background: var(--c-bg);
		border-radius: var(--r-pill);
		padding: 7px 16px;
		margin: 2px 0 4px;
	}
	.composer {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
		padding: 14px 20px;
		border-top: 1px solid var(--c-border);
	}
	.composer textarea {
		flex: 1;
		resize: none;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		padding: 9px 12px;
		color: var(--c-text);
		font-family: var(--f-body);
		font-size: 13px;
	}
	.composer textarea:focus {
		outline: none;
		border-color: var(--c-ink);
	}
	.send-btn {
		padding: 0 18px;
		border-radius: 8px;
		border: none;
		background: var(--c-accent);
		color: #fff;
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.send-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.error-msg {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-error);
		margin: 8px 20px 0;
	}

	@media (max-width: 880px) {
		.back-link {
			display: inline-block;
		}
	}
</style>

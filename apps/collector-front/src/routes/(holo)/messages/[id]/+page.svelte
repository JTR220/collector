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
	import { messages as messagesStore } from '$lib/stores/messages';
	import { fromEventUuid } from '$lib/utils/eventId';
	import GPanel from '$lib/components/galerie/GPanel.svelte';
	import Kicker from '$lib/components/galerie/Kicker.svelte';

	let thread = $state<MessageAPI[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let draft = $state('');
	let sending = $state(false);
	let socket: MessageSocket | null = null;
	let scrollEl: HTMLDivElement | null = null;

	const conversationId = $derived($page.params.id ?? '');
	const myUUID = $derived($auth.user ? toUserUUID($auth.user.id) : '');
	const otherName = $derived(
		thread.length > 0
			? thread[0].sender_id === myUUID
				? thread[0].recipient_name
				: thread[0].sender_name
			: '…'
	);
	const articleRef = $derived(thread.find((m) => m.article_id));
	const articleName = $derived(articleRef?.article_name ?? null);
	const articleId = $derived(articleRef?.article_id ? fromEventUuid(articleRef.article_id) : null);

	async function scrollToBottom() {
		await tick();
		scrollEl?.scrollTo({ top: scrollEl.scrollHeight });
	}

	onMount(async () => {
		if (!$isAuthenticated || !$auth.token) {
			goto('/login');
			return;
		}
		try {
			thread = await fetchConversationMessages($auth.token, conversationId);
			await markConversationRead($auth.token, conversationId);
			messagesStore.refresh($auth.token);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Impossible de charger la conversation.';
		} finally {
			loading = false;
			scrollToBottom();
		}

		socket = connectMessages($auth.token, (msg) => {
			if (msg.conversation_id !== conversationId) return;
			thread = [...thread, msg];
			scrollToBottom();
			if (msg.recipient_id === myUUID && $auth.token) {
				markConversationRead($auth.token, conversationId);
				messagesStore.refresh($auth.token);
			}
		});
	});

	onDestroy(() => socket?.close());

	async function send() {
		const token = $auth.token;
		const body = draft.trim();
		if (!token || !body || thread.length === 0) return;
		const first = thread[0];
		const recipientId = first.sender_id === myUUID ? first.recipient_id : first.sender_id;

		sending = true;
		try {
			const sent = await sendMessage(token, {
				recipientId,
				body,
				articleId: first.article_id,
				articleName: first.article_name
			});
			thread = [...thread, sent];
			draft = '';
			scrollToBottom();
			messagesStore.refresh(token);
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

<a class="back-link" href="/messages">← Conversations</a>

{#if loading}
	<p class="state-msg">Chargement…</p>
{:else if error && thread.length === 0}
	<p class="state-msg error">{error}</p>
{:else}
	<GPanel style="margin-top:10px;display:flex;flex-direction:column;height:60vh">
		<div class="thread-head">
			<Kicker>{otherName}</Kicker>
			{#if articleName && articleId}
				<a class="thread-article" href={`/lot/${articleId}`}>à propos de « {articleName} »</a>
			{/if}
		</div>

		<div class="thread-scroll" bind:this={scrollEl}>
			{#each thread as m (m.id)}
				<div class="msg-row" class:mine={m.sender_id === myUUID}>
					<div class="msg-bubble">
						<span class="msg-text">{m.body}</span>
						<span class="msg-time">{fmtTime(m.created_at)}</span>
					</div>
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
	.state-msg.error {
		color: var(--c-error);
	}
	.back-link {
		display: inline-block;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		text-decoration: none;
	}
	.back-link:hover {
		color: var(--c-ink);
	}
	.thread-head {
		display: flex;
		align-items: baseline;
		gap: 10px;
		padding-bottom: 10px;
		border-bottom: 1px solid var(--c-border);
		flex-shrink: 0;
	}
	.thread-article {
		font-family: var(--f-body);
		font-size: 12px;
		color: var(--c-ink);
		text-decoration: none;
	}
	.thread-article:hover {
		text-decoration: underline;
	}
	.thread-scroll {
		flex: 1;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px 4px;
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
	.msg-text {
		font-family: var(--f-body);
		font-size: 13.5px;
		color: var(--c-text);
		white-space: pre-wrap;
		word-break: break-word;
	}
	.msg-time {
		font-family: var(--f-body);
		font-size: 9.5px;
		color: var(--c-text-muted);
		align-self: flex-end;
	}
	.composer {
		display: flex;
		gap: 8px;
		padding-top: 10px;
		border-top: 1px solid var(--c-border);
		flex-shrink: 0;
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
		margin: 8px 0 0;
	}
</style>

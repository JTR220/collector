<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { notifications } from '$lib/stores/notifications';

	let open = $state(false);

	const typeIcons: Record<string, string> = {
		PRICE_DROP: '▼',
		PRICE_SPIKE: '▲',
		FRAUD_ALERT: '⚠',
		NEW_ITEM: '◆',
		ITEM_SOLD: '✓'
	};

	function toggle() {
		open = !open;
	}

	function onRead(id: string) {
		if ($auth.token) notifications.markRead($auth.token, id);
	}

	function onReadAll() {
		if ($auth.token) notifications.markAllRead($auth.token);
	}

	function timeAgo(iso: string): string {
		const diff = Date.now() - new Date(iso).getTime();
		const min = Math.floor(diff / 60000);
		if (min < 1) return "a l'instant";
		if (min < 60) return `il y a ${min} min`;
		const h = Math.floor(min / 60);
		if (h < 24) return `il y a ${h} h`;
		return `il y a ${Math.floor(h / 24)} j`;
	}
</script>

<div class="bell-wrap">
	<button class="bell-btn" onclick={toggle} aria-label="Notifications" title="Notifications">
		<span class="bell-icon">◎</span>
		{#if $notifications.unreadCount > 0}
			<span class="badge"
				>{$notifications.unreadCount > 99 ? '99+' : $notifications.unreadCount}</span
			>
		{/if}
	</button>

	{#if open}
		<div class="dropdown">
			<div class="dropdown-head">
				<span class="dropdown-title">Notifications</span>
				{#if $notifications.unreadCount > 0}
					<button class="readall-btn" onclick={onReadAll}>Tout marquer lu</button>
				{/if}
			</div>
			{#if $notifications.items.length === 0}
				<p class="empty">Aucune notification</p>
			{:else}
				<ul class="list">
					{#each $notifications.items.slice(0, 12) as notif (notif.id)}
						<li>
							<button class="item" class:item-unread={!notif.read} onclick={() => onRead(notif.id)}>
								<span class="item-icon">{typeIcons[notif.type] ?? '◆'}</span>
								<span class="item-body">
									<span class="item-title">{notif.title}</span>
									{#if notif.body}<span class="item-text">{notif.body}</span>{/if}
									<span class="item-time">{timeAgo(notif.created_at)}</span>
								</span>
								{#if !notif.read}<span class="dot"></span>{/if}
							</button>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</div>

<style>
	.bell-wrap {
		position: relative;
	}
	.bell-btn {
		position: relative;
		background: none;
		border: 1px solid rgba(236, 229, 218, 0.1);
		border-radius: 6px;
		padding: 5px 9px;
		cursor: pointer;
		transition:
			color 120ms,
			border-color 120ms;
	}
	.bell-icon {
		font-size: 14px;
		color: #a39a8c;
	}
	.bell-btn:hover .bell-icon {
		color: #ece5da;
	}
	.bell-btn:hover {
		border-color: rgba(236, 229, 218, 0.22);
	}
	.badge {
		position: absolute;
		top: -7px;
		right: -7px;
		min-width: 16px;
		padding: 1px 4px;
		border-radius: 999px;
		background: #86b3a4;
		color: #191714;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 9px;
		font-weight: 700;
		text-align: center;
	}
	.dropdown {
		position: absolute;
		top: calc(100% + 10px);
		right: 0;
		width: 320px;
		max-height: 420px;
		overflow-y: auto;
		background: #211e1a;
		border: 1px solid rgba(236, 229, 218, 0.14);
		border-radius: 10px;
		box-shadow: 0 12px 32px -8px rgba(0, 0, 0, 0.7);
		z-index: 10000;
	}
	.dropdown-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 14px;
		border-bottom: 1px solid rgba(236, 229, 218, 0.08);
	}
	.dropdown-title {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 10.5px;
		font-weight: 600;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #766d60;
	}
	.readall-btn {
		background: none;
		border: none;
		font-size: 11px;
		color: #86b3a4;
		cursor: pointer;
	}
	.readall-btn:hover {
		text-decoration: underline;
	}
	.empty {
		padding: 24px 14px;
		text-align: center;
		font-size: 12px;
		color: #766d60;
	}
	.list {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.item {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		width: 100%;
		padding: 11px 14px;
		background: none;
		border: none;
		border-bottom: 1px solid rgba(236, 229, 218, 0.05);
		text-align: left;
		cursor: pointer;
		transition: background 120ms;
	}
	.item:hover {
		background: rgba(134, 179, 164, 0.06);
	}
	.item-unread {
		background: rgba(134, 179, 164, 0.04);
	}
	.item-icon {
		font-size: 13px;
		color: #86b3a4;
		flex-shrink: 0;
	}
	.item-body {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}
	.item-title {
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 12.5px;
		font-weight: 600;
		color: #ece5da;
	}
	.item-text {
		font-size: 11px;
		color: #a39a8c;
		line-height: 1.4;
	}
	.item-time {
		font-size: 10px;
		color: #766d60;
	}
	.dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #86b3a4;
		flex-shrink: 0;
		margin-top: 4px;
	}
</style>

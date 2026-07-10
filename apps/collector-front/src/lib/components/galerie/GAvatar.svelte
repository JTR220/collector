<script lang="ts">
	let {
		initials = '',
		size = 40,
		square = false
	}: { initials?: string; size?: number; square?: boolean } = $props();
</script>

<div
	class="g-avatar"
	style="
		width:{size}px;
		height:{size}px;
		border-radius:{square ? size * 0.28 : 50}%;
		font-size:{size * 0.4}px;
	"
>
	{initials}
</div>

<style>
	.g-avatar {
		flex-shrink: 0;
		border: none;
		background: var(--c-ink);
		color: var(--c-bg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: var(--f-body);
		font-weight: 600;
	}
</style>

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
		border-radius:{square ? size*0.28 : 50}%;
		font-size:{size*0.4}px;
	"
>
	{initials}
</div>

<style>
	.g-avatar {
		flex-shrink: 0;
		border: 1px solid rgba(236,229,218,0.16);
		background: rgba(255,255,255,0.05);
		color: #86b3a4;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: 'Newsreader', Georgia, serif;
	}
</style>

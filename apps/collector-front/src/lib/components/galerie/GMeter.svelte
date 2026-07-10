<script lang="ts">
	let {
		value = 0,
		height = 4,
		color = '#1e3b2c'
	}: { value?: number; height?: number; color?: string } = $props();
	const w = $derived(Math.max(0, Math.min(100, value)));
</script>

<div class="track" style="height:{height}px;border-radius:{height / 2}px">
	<div class="fill" style="width:{w}%;background:{color};border-radius:{height / 2}px"></div>
</div>

<style>
	.track {
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		overflow: hidden;
	}
	.fill {
		height: 100%;
		opacity: 0.85;
		transition: width 300ms ease;
	}
</style>

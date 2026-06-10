<script lang="ts">
	let { value = 0, height = 4, color = '#86b3a4' }: { value?: number; height?: number; color?: string } = $props();
	const w = $derived(Math.max(0, Math.min(100, value)));
</script>

<div class="track" style="height:{height}px;border-radius:{height/2}px">
	<div class="fill" style="width:{w}%;background:{color};border-radius:{height/2}px"></div>
</div>

<style>
	.track {
		background: rgba(255,255,255,0.05);
		border: 1px solid rgba(236,229,218,0.10);
		overflow: hidden;
	}
	.fill {
		height: 100%;
		opacity: 0.6;
		transition: width 300ms ease;
	}
</style>

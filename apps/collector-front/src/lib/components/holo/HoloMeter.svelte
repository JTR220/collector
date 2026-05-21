<script lang="ts">
	type Props = { value: number; hue?: number; height?: number };
	let { value, hue = 215, height = 8 }: Props = $props();
</script>

<div class="track" style="height:{height}px">
	<div
		class="fill"
		style="width:{Math.min(100, value)}%"
	></div>
</div>

<style>
	.track {
		width: 100%;
		background: rgba(255,255,255,0.06);
		border: 1px solid rgba(255,255,255,0.07);
		border-radius: 2px;
		overflow: hidden;
	}
	.fill {
		height: 100%;
		border-radius: 2px;
		background: linear-gradient(90deg, #a8c8e4, #6a7280);
		box-shadow: 0 0 12px rgba(168,200,228,0.35);
		transition: width 600ms ease;
	}
</style>

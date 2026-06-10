<script lang="ts">
	type Props = { item: { glyph: string }; size?: number };
	let { item, size = 140 }: Props = $props();
</script>

<div class="art" style="width:{size}px;height:{size}px">
	<div class="scanlines" aria-hidden="true"></div>
	<span class="glyph">{item.glyph}</span>
</div>

<style>
	.art {
		position: relative;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
		background: radial-gradient(
			120% 90% at 30% 20%,
			oklch(0.55 0.08 215) 0%,
			oklch(0.32 0.06 215) 55%,
			oklch(0.18 0.04 215) 100%
		);
		box-shadow: inset 0 0 40px rgba(0, 0, 0, 0.4);
	}
	.scanlines {
		pointer-events: none;
		position: absolute;
		inset: 0;
		background: repeating-linear-gradient(
			to bottom,
			rgba(255, 255, 255, 0.07) 0 1px,
			transparent 1px 3px
		);
		mix-blend-mode: overlay;
		opacity: 0.5;
	}
	.glyph {
		position: relative;
		z-index: 1;
		font-family: 'Newsreader', Georgia, serif;
		font-size: 48px;
		color: rgba(255, 255, 255, 0.94);
		text-shadow: 0 4px 22px rgba(0, 0, 0, 0.45);
	}
</style>

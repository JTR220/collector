<script lang="ts">
	type Props = {
		glow?: string;
		style?: string;
		children?: import('svelte').Snippet;
	};
	let { glow, style = '', children }: Props = $props();

	const shadow = glow
		? `0 0 0 1px ${glow}33, 0 20px 50px -28px ${glow}66, inset 0 0 0 1px rgba(255,255,255,0.07)`
		: `inset 0 0 0 1px rgba(255,255,255,0.07)`;
</script>

<div class="panel" style="box-shadow:{shadow};{style}">
	{@render children?.()}
</div>

<style>
	.panel {
		background: linear-gradient(180deg, #181a20 0%, #0d0f13 100%);
		border: 1px solid rgba(255,255,255,0.07);
		border-radius: 12px;
		padding: 18px;
		position: relative;
		overflow: hidden;
	}
	.panel::after {
		content: '';
		pointer-events: none;
		position: absolute;
		inset: 0;
		background: repeating-linear-gradient(
			to bottom,
			rgba(255,255,255,0.025) 0 1px,
			transparent 1px 3px
		);
		mix-blend-mode: overlay;
		opacity: 0.4;
	}
</style>

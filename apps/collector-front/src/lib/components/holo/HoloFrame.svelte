<script lang="ts">
	let { children }: { children?: import('svelte').Snippet } = $props();
</script>

<div class="steel-frame">
	<div class="steel-scanlines" aria-hidden="true"></div>
	<div class="steel-noise" aria-hidden="true"></div>
	{@render children?.()}
</div>

<style>
	.steel-frame {
		position: relative;
		min-height: 100vh;
		background:
			radial-gradient(60% 50% at 20% 0%, rgba(170, 196, 224, 0.08) 0%, transparent 60%),
			radial-gradient(50% 50% at 90% 100%, rgba(120, 140, 160, 0.05) 0%, transparent 55%), #0e1014;
		color: #e8eaed;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		padding: 22px 30px 24px;
		overflow-x: hidden;
	}
	.steel-scanlines {
		pointer-events: none;
		position: fixed;
		inset: 0;
		z-index: 9999;
		background: repeating-linear-gradient(
			to bottom,
			rgba(255, 255, 255, 0.04) 0 1px,
			transparent 1px 3px
		);
		mix-blend-mode: overlay;
		opacity: 0.22;
	}
	.steel-noise {
		pointer-events: none;
		position: fixed;
		inset: 0;
		z-index: 9998;
		opacity: 0.06;
		mix-blend-mode: overlay;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='n'%3E%3CfeTurbulence baseFrequency='0.9' numOctaves='2'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.5'/%3E%3C/svg%3E");
	}
</style>

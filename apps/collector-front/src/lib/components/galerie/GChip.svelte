<script lang="ts">
	let { children, color = '' }: { children?: import('svelte').Snippet; color?: string } = $props();
</script>

<span
	class="g-chip"
	style={color ? `color:${color};border-color:${color}44;background:${color}12` : ''}
>
	{@render children?.()}
</span>

<style>
	.g-chip {
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		font-size: 10.5px;
		letter-spacing: 0.02em;
		color: #a39a8c;
		padding: 3px 8px;
		border-radius: 4px;
		border: 1px solid rgba(236, 229, 218, 0.1);
		background: rgba(255, 255, 255, 0.05);
		white-space: nowrap;
	}
</style>

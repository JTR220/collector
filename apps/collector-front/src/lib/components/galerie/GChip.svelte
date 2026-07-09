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
		font-family: var(--f-body);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.02em;
		color: var(--c-ink);
		padding: 3px 8px;
		border-radius: 5px;
		border: none;
		background: var(--c-badge-verified-bg);
		white-space: nowrap;
	}
</style>

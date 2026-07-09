<script lang="ts">
	let { children, style = '' }: { children?: import('svelte').Snippet; style?: string } = $props();
</script>

<div class="g-panel" {style}>
	{@render children?.()}
</div>

<style>
	.g-panel {
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: var(--r-card);
		padding: 20px;
	}
</style>

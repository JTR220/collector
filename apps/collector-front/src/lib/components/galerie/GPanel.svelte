<script lang="ts">
	let { children, style = '' }: { children?: import('svelte').Snippet; style?: string } = $props();
</script>

<div class="g-panel" {style}>
	{@render children?.()}
</div>

<style>
	.g-panel {
		background: #221f1b;
		border: 1px solid rgba(236,229,218,0.10);
		border-radius: 9px;
		padding: 18px;
	}
</style>

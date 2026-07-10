<script lang="ts">
	let { children }: { children?: import('svelte').Snippet } = $props();
</script>

<div class="g-frame">
	{@render children?.()}
</div>

<style>
	.g-frame {
		width: 100%;
		min-height: 100vh;
		background: var(--c-bg);
		color: var(--c-text);
		font-family: var(--f-body);
		overflow-x: hidden;
		box-sizing: border-box;
	}
</style>

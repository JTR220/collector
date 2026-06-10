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
		background: #191714;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		padding: 26px 38px 40px;
		overflow-x: hidden;
		box-sizing: border-box;
	}
</style>

<script lang="ts">
	type Props = { glow?: string; size?: number; children?: import('svelte').Snippet };
	let { glow = 'rgba(168,200,228,0.6)', size = 48, children }: Props = $props();
</script>

<h1 class="title" style="font-size:{size}px;text-shadow:0 0 28px {glow}">
	{@render children?.()}
</h1>

<style>
	.title {
		font-family: 'Newsreader', Georgia, serif;
		line-height: 0.95;
		margin: 0;
		letter-spacing: -0.02em;
		color: #e8eaed;
	}
</style>

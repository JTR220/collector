<script lang="ts">
	type Props = { color?: string; children?: import('svelte').Snippet };
	let { color = '#8a909a', children }: Props = $props();
</script>

<span class="chip" style="border-color:{color}55;background:{color}14;color:{color}">
	{@render children?.()}
</span>

<style>
	.chip {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		border: 1px solid;
		border-radius: 4px;
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.18em;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
	}
</style>

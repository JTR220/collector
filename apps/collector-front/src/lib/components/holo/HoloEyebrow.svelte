<script lang="ts">
	type Props = { color?: string; children?: import('svelte').Snippet };
	let { color = '#8a909a', children }: Props = $props();
</script>

<div class="eyebrow" style="color:{color}">{@render children?.()}</div>

<style>
	.eyebrow {
		font-size: 11px;
		letter-spacing: 0.34em;
		margin-bottom: 10px;
		font-family: 'IBM Plex Mono', ui-monospace, monospace;
		text-transform: uppercase;
	}
</style>

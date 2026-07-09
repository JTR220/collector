<script lang="ts">
	let {
		children,
		color = '#8a7a64',
		style = ''
	}: { children?: import('svelte').Snippet; color?: string; style?: string } = $props();
</script>

<div class="kicker" style="color:{color};{style}">{@render children?.()}</div>

<style>
	.kicker {
		font-family: var(--f-body);
		font-size: 11px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		font-weight: 600;
	}
</style>

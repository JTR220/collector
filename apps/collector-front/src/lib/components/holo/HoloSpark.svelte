<script lang="ts">
	import { sparkPath } from '$lib/utils/format';
	type Props = { values: number[]; color?: string; width?: number; height?: number };
	let { values, color = '#a8c8e4', width = 140, height = 32 }: Props = $props();

	const path = $derived(sparkPath(values, width, height));
</script>

<svg {width} {height} style="display:block;overflow:visible">
	<path
		d={path}
		fill="none"
		stroke={color}
		stroke-width="3"
		opacity="0.2"
		style="filter:blur(3px)"
	/>
	<path d={path} fill="none" stroke={color} stroke-width="1.5" opacity="0.85" />
</svg>

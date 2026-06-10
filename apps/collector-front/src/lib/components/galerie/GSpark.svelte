<script lang="ts">
	import { sparkPath } from '$lib/utils/format';

	let {
		values = [],
		color = '#86b3a4',
		w = 86,
		h = 26,
		dot = true
	}: { values?: number[]; color?: string; w?: number; h?: number; dot?: boolean } = $props();

	const pad = 3;
	const path = $derived(sparkPath(values, w, h));

	const lastDot = $derived((): { x: number; y: number } | null => {
		if (!values.length) return null;
		const min = Math.min(...values), max = Math.max(...values);
		const range = max - min || 1;
		const last = values[values.length - 1];
		return {
			x: w - pad,
			y: pad + (h - pad * 2) * (1 - (last - min) / range)
		};
	});
</script>

<svg width={w} height={h} style="display:block;overflow:visible">
	<path d={path} fill="none" stroke={color} stroke-width="1.4"
		stroke-linejoin="round" stroke-linecap="round" opacity="0.9" />
	{#if dot && lastDot()}
		<circle cx={lastDot()!.x} cy={lastDot()!.y} r="2.1" fill={color} />
	{/if}
</svg>

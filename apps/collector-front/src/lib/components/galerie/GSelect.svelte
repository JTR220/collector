<script lang="ts">
	// Dropdown thémé (holo sombre) en remplacement du <select> natif, dont la
	// liste déroulante n'est pas stylable de façon fiable selon les navigateurs.
	type Option = { value: string | number; label: string };

	let {
		value = $bindable(),
		options,
		placeholder = 'Choisir…',
		ariaLabel = '',
		compact = false
	}: {
		value: string | number;
		options: Option[];
		placeholder?: string;
		ariaLabel?: string;
		compact?: boolean;
	} = $props();

	let open = $state(false);
	let root: HTMLDivElement;

	const selected = $derived(options.find((o) => o.value === value));
	const activeIndex = $derived(Math.max(0, options.findIndex((o) => o.value === value)));

	function choose(o: Option) {
		value = o.value;
		open = false;
	}

	function onKey(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			return;
		}
		if (!open && (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ')) {
			e.preventDefault();
			open = true;
			return;
		}
		if (open && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
			e.preventDefault();
			const dir = e.key === 'ArrowDown' ? 1 : -1;
			const next = Math.min(options.length - 1, Math.max(0, activeIndex + dir));
			value = options[next].value;
		}
	}

	function onWindowClick(e: MouseEvent) {
		if (open && root && !root.contains(e.target as Node)) open = false;
	}
</script>

<svelte:window onclick={onWindowClick} />

<div class="gs" class:gs-compact={compact} bind:this={root}>
	<button
		type="button"
		class="gs-btn"
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-label={ariaLabel}
		onclick={() => (open = !open)}
		onkeydown={onKey}
	>
		<span class="gs-value" class:gs-placeholder={!selected}>{selected?.label ?? placeholder}</span>
		<span class="gs-chev" class:gs-chev-open={open} aria-hidden="true">▾</span>
	</button>

	{#if open}
		<ul class="gs-list" role="listbox" aria-label={ariaLabel}>
			{#each options as o (o.value)}
				<li
					role="option"
					aria-selected={o.value === value}
					class="gs-opt"
					class:gs-opt-active={o.value === value}
					onclick={() => choose(o)}
				>
					{o.label}
					{#if o.value === value}<span class="gs-check" aria-hidden="true">✓</span>{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.gs {
		position: relative;
		display: block;
		width: 100%;
	}
	.gs-btn {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		width: 100%;
		min-width: 150px;
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid rgba(236, 229, 218, 0.12);
		border-radius: 8px;
		padding: 10px 12px;
		color: #ece5da;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		cursor: pointer;
		outline: none;
		transition:
			border-color 150ms,
			box-shadow 150ms;
	}
	.gs-compact .gs-btn {
		min-width: 0;
	}
	.gs-btn:hover {
		border-color: rgba(236, 229, 218, 0.22);
	}
	.gs-btn:focus-visible,
	.gs-btn[aria-expanded='true'] {
		border-color: rgba(134, 179, 164, 0.5);
		box-shadow: 0 0 0 3px rgba(134, 179, 164, 0.08);
	}
	.gs-value {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.gs-placeholder {
		color: rgba(236, 229, 218, 0.4);
	}
	.gs-chev {
		color: #86b3a4;
		font-size: 11px;
		transition: transform 150ms;
		flex-shrink: 0;
	}
	.gs-chev-open {
		transform: rotate(180deg);
	}

	.gs-list {
		position: absolute;
		z-index: 40;
		top: calc(100% + 6px);
		left: 0;
		right: 0;
		margin: 0;
		padding: 5px;
		list-style: none;
		max-height: 260px;
		overflow-y: auto;
		background: #201d19;
		border: 1px solid rgba(236, 229, 218, 0.14);
		border-radius: 9px;
		box-shadow: 0 12px 30px rgba(0, 0, 0, 0.45);
	}
	.gs-opt {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 9px 11px;
		border-radius: 6px;
		color: #cbc3b6;
		font-family: 'Hanken Grotesk', system-ui, sans-serif;
		font-size: 13px;
		cursor: pointer;
		transition:
			background 120ms,
			color 120ms;
	}
	.gs-opt:hover {
		background: rgba(255, 255, 255, 0.05);
		color: #ece5da;
	}
	.gs-opt-active {
		color: #86b3a4;
		background: rgba(134, 179, 164, 0.08);
	}
	.gs-check {
		color: #86b3a4;
		font-size: 11px;
		flex-shrink: 0;
	}
</style>

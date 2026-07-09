<script lang="ts">
	// Modale de confirmation themee (remplace window.confirm), avec case a
	// cocher optionnelle pour les actions qui le necessitent (ex: suppression
	// groupee ou confirmation explicite d'une action destructive).
	let {
		open = $bindable(false),
		title = 'Confirmer',
		message = '',
		confirmLabel = 'Confirmer',
		cancelLabel = 'Annuler',
		danger = false,
		checkboxLabel = '',
		checkboxChecked = $bindable(false),
		onConfirm,
		onCancel
	}: {
		open?: boolean;
		title?: string;
		message?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		danger?: boolean;
		checkboxLabel?: string;
		checkboxChecked?: boolean;
		onConfirm?: () => void;
		onCancel?: () => void;
	} = $props();

	function confirm() {
		open = false;
		onConfirm?.();
	}

	function cancel() {
		open = false;
		onCancel?.();
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') cancel();
	}

	// Si une case a cocher est fournie, elle doit etre cochee pour activer la confirmation.
	const confirmDisabled = $derived(checkboxLabel !== '' && !checkboxChecked);
</script>

<svelte:window onkeydown={open ? onKeydown : undefined} />

{#if open}
	<div
		class="gc-overlay"
		onclick={cancel}
		onkeydown={onKeydown}
		role="button"
		tabindex="-1"
		aria-label={cancelLabel}
	>
		<div
			class="gc-modal"
			role="alertdialog"
			aria-modal="true"
			aria-labelledby="gc-title"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
		>
			<h2 id="gc-title" class="gc-title">{title}</h2>
			{#if message}<p class="gc-message">{message}</p>{/if}

			{#if checkboxLabel}
				<label class="gc-checkbox">
					<input class="chk" type="checkbox" bind:checked={checkboxChecked} />
					<span>{checkboxLabel}</span>
				</label>
			{/if}

			<div class="gc-actions">
				<button type="button" class="gc-btn gc-btn-cancel" onclick={cancel}>
					{cancelLabel}
				</button>
				<button
					type="button"
					class="gc-btn gc-btn-confirm"
					class:gc-btn-danger={danger}
					disabled={confirmDisabled}
					onclick={confirm}
				>
					{confirmLabel}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.gc-overlay {
		position: fixed;
		inset: 0;
		background: rgba(43, 38, 32, 0.45);
		backdrop-filter: blur(2px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
		padding: 24px;
		box-sizing: border-box;
	}

	.gc-modal {
		width: 100%;
		max-width: 380px;
		background: var(--c-surface);
		border: 1px solid var(--c-border);
		border-radius: 12px;
		padding: 24px;
		box-shadow: 0 20px 50px rgba(43, 38, 32, 0.25);
		animation: gc-in 140ms ease-out;
	}

	@keyframes gc-in {
		from {
			opacity: 0;
			transform: translateY(6px) scale(0.98);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	.gc-title {
		font-family: var(--f-serif);
		font-weight: 600;
		font-size: 22px;
		color: var(--c-text);
		margin: 0 0 8px;
	}

	.gc-message {
		font-family: var(--f-body);
		font-size: 13.5px;
		line-height: 1.5;
		color: var(--c-text-muted);
		margin: 0 0 18px;
	}

	.gc-checkbox {
		display: flex;
		align-items: center;
		gap: 9px;
		padding: 10px 12px;
		margin-bottom: 20px;
		background: var(--c-bg);
		border: 1px solid var(--c-border);
		border-radius: 8px;
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text);
		cursor: pointer;
	}
	/* Case a cocher themee (remplace l'apparence native du navigateur) */
	.chk {
		appearance: none;
		-webkit-appearance: none;
		width: 16px;
		height: 16px;
		margin: 0;
		border-radius: 4px;
		border: 1px solid var(--c-border);
		background: var(--c-surface);
		cursor: pointer;
		position: relative;
		flex-shrink: 0;
		transition:
			background 120ms,
			border-color 120ms;
	}
	.chk:hover {
		border-color: var(--c-ink);
	}
	.chk:checked {
		background: var(--c-ink);
		border-color: var(--c-ink);
	}
	.chk:checked::after {
		content: '';
		position: absolute;
		left: 5px;
		top: 2px;
		width: 4px;
		height: 8px;
		border: solid var(--c-bg);
		border-width: 0 2px 2px 0;
		transform: rotate(45deg);
	}
	.chk:focus-visible {
		outline: none;
		box-shadow: 0 0 0 3px rgba(30, 59, 44, 0.2);
	}

	.gc-actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
	}

	.gc-btn {
		padding: 10px 18px;
		border-radius: 7px;
		font-family: var(--f-body);
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition:
			filter 120ms,
			border-color 120ms,
			background 120ms;
	}

	.gc-btn-cancel {
		background: none;
		border: 1px solid var(--c-border);
		color: var(--c-text-tertiary);
	}
	.gc-btn-cancel:hover {
		color: var(--c-text);
		border-color: var(--c-ink);
	}

	.gc-btn-confirm {
		border: none;
		background: var(--c-accent);
		color: #fff;
	}
	.gc-btn-confirm:hover:not(:disabled) {
		filter: brightness(1.08);
	}
	.gc-btn-confirm:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}

	.gc-btn-danger {
		background: var(--c-error);
		color: #fff;
	}
</style>

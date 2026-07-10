<footer class="g-footer">
	<span class="g-footer-brand">Collector<span class="g-footer-dot">.</span>shop</span>
	<nav class="g-footer-links">
		<a href="/confidentialite">Confidentialité</a>
		<a href="/cgu">CGU</a>
	</nav>
</footer>

<style>
	.g-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 20px 48px;
		border-top: 1px solid var(--c-border);
		flex-wrap: wrap;
	}
	.g-footer-brand {
		font-family: var(--f-serif);
		font-size: 14px;
		font-weight: 700;
		color: var(--c-text-muted);
	}
	.g-footer-dot {
		color: var(--c-accent);
	}
	.g-footer-links {
		display: flex;
		gap: 20px;
	}
	.g-footer-links a {
		font-family: var(--f-body);
		font-size: 12.5px;
		color: var(--c-text-tertiary);
		text-decoration: none;
		transition: color 120ms;
	}
	.g-footer-links a:hover {
		color: var(--c-ink);
	}
	@media (max-width: 640px) {
		.g-footer {
			padding: 16px 20px;
		}
	}
</style>

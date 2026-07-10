<div class="empty-wrap">
	<div class="empty-state">
		<span class="empty-glyph">✉</span>
		<p class="empty-title">Sélectionnez une conversation</p>
		<p class="empty-text">
			Choisissez un échange dans la liste à gauche, ou contactez un vendeur depuis une fiche produit
			pour démarrer une nouvelle conversation.
		</p>
	</div>
</div>

<style>
	.empty-wrap {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.empty-state {
		text-align: center;
		max-width: 320px;
		padding: 40px 20px;
	}
	.empty-glyph {
		display: inline-block;
		font-size: 28px;
		color: var(--c-icon-muted);
		margin-bottom: 10px;
	}
	.empty-title {
		font-family: var(--f-serif);
		font-size: 18px;
		font-weight: 600;
		color: var(--c-text);
		margin: 0 0 6px;
	}
	.empty-text {
		font-family: var(--f-body);
		font-size: 13px;
		color: var(--c-text-muted);
		line-height: 1.6;
		margin: 0;
	}
</style>

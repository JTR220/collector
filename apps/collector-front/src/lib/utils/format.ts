export const eur = (n: number) =>
	new Intl.NumberFormat('fr-FR', {
		style: 'currency',
		currency: 'EUR',
		maximumFractionDigits: 0
	}).format(n);

export const eurC = (n: number) =>
	new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(n);

export const pct = (n: number) => (n >= 0 ? '+' : '') + n.toFixed(1) + '%';

export function sparkPath(values: number[], w: number, h: number, pad = 4): string {
	if (!values || !values.length) return '';
	const min = Math.min(...values);
	const max = Math.max(...values);
	const range = max - min || 1;
	const step = (w - pad * 2) / (values.length - 1);
	return values
		.map((v, i) => {
			const x = pad + i * step;
			const y = pad + (h - pad * 2) * (1 - (v - min) / range);
			return (i === 0 ? 'M' : 'L') + x.toFixed(1) + ',' + y.toFixed(1);
		})
		.join(' ');
}

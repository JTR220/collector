export type Collectible = {
	id: string;
	name: string;
	series: string;
	category: 'TCG' | 'Console' | 'Comics' | 'Vinyle' | 'Designer Toy' | 'Horlogerie';
	year: number;
	rarity: string;
	rarityScore: 1 | 2 | 3 | 4 | 5;
	grade: string;
	price: number;
	shipping: number;
	seller: string;
	sellerScore: number;
	delta: number;
	history: number[];
	hue: number;
	accent: string;
	swatch: [string, string, string];
	glyph: string;
};

export const COLLECTIBLES: Collectible[] = [
	{
		id: 'PKM-001', name: 'Charizard', series: 'Base Set, 1ère édition', category: 'TCG',
		year: 1999, rarity: 'Holo Rare', rarityScore: 5, grade: 'PSA 9',
		price: 18400, shipping: 24, seller: 'kanto_archive', sellerScore: 4.98, delta: 6.2,
		history: [12000,12800,13900,13200,14600,15800,17100,18400], hue: 18,
		accent: '#ff7a1a', swatch: ['#ff7a1a','#ffd06b','#1a1208'], glyph: '卡'
	},
	{
		id: 'GBC-014', name: 'Game Boy Color', series: 'Édition Pikachu, scellé', category: 'Console',
		year: 1998, rarity: 'Sealed', rarityScore: 5, grade: 'Mint',
		price: 1290, shipping: 18, seller: 'tokyo_loop', sellerScore: 4.91, delta: 1.8,
		history: [820,880,920,1010,1080,1120,1240,1290], hue: 48,
		accent: '#ffd84a', swatch: ['#ffd84a','#1e90c2','#0a1a22'], glyph: '電'
	},
	{
		id: 'CMX-007', name: 'Action Comics #1', series: 'DC, reprint 1988', category: 'Comics',
		year: 1988, rarity: 'Near Mint', rarityScore: 4, grade: 'CGC 9.6',
		price: 640, shipping: 14, seller: 'panel_press', sellerScore: 4.84, delta: -2.1,
		history: [710,705,680,700,685,660,650,640], hue: 220,
		accent: '#3a6ef0', swatch: ['#3a6ef0','#e63946','#0b0b1a'], glyph: 'S'
	},
	{
		id: 'VNL-022', name: 'Daft Punk — Discovery', series: 'Vinyle, 1ère presse 2001', category: 'Vinyle',
		year: 2001, rarity: 'Rare', rarityScore: 4, grade: 'VG+',
		price: 320, shipping: 12, seller: 'groove_atlas', sellerScore: 4.96, delta: 3.4,
		history: [220,240,250,265,280,295,310,320], hue: 350,
		accent: '#ff3d8b', swatch: ['#ff3d8b','#f6c945','#190a14'], glyph: '♪'
	},
	{
		id: 'FIG-101', name: 'Bearbrick 1000%', series: 'Andy Warhol, 2022', category: 'Designer Toy',
		year: 2022, rarity: 'Limited', rarityScore: 3, grade: 'MIB',
		price: 1180, shipping: 32, seller: 'soho_pulse', sellerScore: 4.79, delta: 0.4,
		history: [1100,1140,1130,1160,1170,1150,1175,1180], hue: 280,
		accent: '#b35bff', swatch: ['#b35bff','#5ee6c8','#120a1a'], glyph: '★'
	},
	{
		id: 'WAT-045', name: 'Casio F-91W', series: 'Mod custom NATO bleu', category: 'Horlogerie',
		year: 1991, rarity: 'Common', rarityScore: 2, grade: 'EX',
		price: 89, shipping: 8, seller: 'midnight_wrist', sellerScore: 4.65, delta: -0.6,
		history: [85,90,88,92,91,87,90,89], hue: 195,
		accent: '#3fd1ff', swatch: ['#3fd1ff','#1a2630','#04080c'], glyph: '◷'
	}
];

<script lang="ts">
    import { fly, fade } from 'svelte/transition';

    type PricePoint = { time: string; price: number };
    type AlertType = 'SPIKE' | 'DUMP' | 'STABLE';

    const items = [
        { id: '1', title: 'Charizard 1ère éd.',      basePrice: 350 },
        { id: '2', title: 'Black Lotus Magic',         basePrice: 800 },
        { id: '3', title: 'Game Boy Color',            basePrice: 120 },
        { id: '4', title: 'Vinyle Daft Punk',          basePrice: 65  },
    ];

    let selectedItem = $state(items[0]);
    let history = $state<PricePoint[]>([]);
    let alert = $state<AlertType>('STABLE');
    let lastPrice = $state(0);
    let variation = $state(0);

    // Génère un historique initial
    function generateHistory(base: number): PricePoint[] {
        return Array.from({ length: 20 }, (_, i) => {
            const noise = (Math.random() - 0.5) * base * 0.15;
            return {
                time: `${String(i).padStart(2, '0')}:00`,
                price: Math.round(base + noise),
            };
        });
    }

    function detectAlert(prices: number[]): AlertType {
        const last = prices[prices.length - 1];
        const prev = prices[prices.length - 5] ?? prices[0];
        const change = (last - prev) / prev;
        if (change > 0.15) return 'SPIKE';
        if (change < -0.15) return 'DUMP';
        return 'STABLE';
    }

    function selectItem(item: typeof items[0]) {
        selectedItem = item;
        history = generateHistory(item.basePrice);
        lastPrice = history[history.length - 1].price;
        alert = detectAlert(history.map(h => h.price));
        variation = Math.round(((lastPrice - item.basePrice) / item.basePrice) * 100);
    }

    // SVG path helpers
    const W = 600;
    const H = 200;
    const PAD = 20;

    let svgPath = $derived.by(() => {
        if (history.length < 2) return '';
        const prices = history.map(h => h.price);
        const min = Math.min(...prices);
        const max = Math.max(...prices);
        const range = max - min || 1;

        const points = prices.map((p, i) => {
            const x = PAD + (i / (prices.length - 1)) * (W - PAD * 2);
            const y = H - PAD - ((p - min) / range) * (H - PAD * 2);
            return `${x},${y}`;
        });

        return `M ${points.join(' L ')}`;
    });

    let fillPath = $derived(svgPath
        ? `${svgPath} L ${W - PAD},${H - PAD} L ${PAD},${H - PAD} Z`
        : ''
    );

    const alertConfig = {
        SPIKE: { label: '🚨 SPIKE détecté — hausse anormale', color: 'text-red-400',   bg: 'bg-red-400/10   border-red-400/30'   },
        DUMP:  { label: '📉 DUMP détecté — baisse suspecte',  color: 'text-blue-400',  bg: 'bg-blue-400/10  border-blue-400/30'  },
        STABLE:{ label: '✅ Prix stable',                     color: 'text-green-400', bg: 'bg-green-400/10 border-green-400/30' },
    };

    // Init
    selectItem(items[0]);

    // Simule des updates temps réel
    $effect(() => {
        const interval = setInterval(() => {
            const noise = (Math.random() - 0.5) * selectedItem.basePrice * 0.08;
            const newPrice = Math.round((history[history.length - 1]?.price ?? selectedItem.basePrice) + noise);
            const newPoint: PricePoint = {
                time: new Date().toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' }),
                price: newPrice,
            };
            history = [...history.slice(-29), newPoint];
            lastPrice = newPrice;
            alert = detectAlert(history.map(h => h.price));
            variation = Math.round(((lastPrice - selectedItem.basePrice) / selectedItem.basePrice) * 100);
        }, 1500);

        return () => clearInterval(interval);
    });
</script>

<div class="min-h-screen bg-slate-950 text-white">

    <!-- Header -->
    <div class="border-b border-white/10 px-6 py-4 flex items-center justify-between">
        <a href="/" class="text-2xl font-black">collector<span class="text-amber-400">.shop</span></a>
        <span class="text-slate-400 text-sm">Price Tracker</span>
    </div>

    <div class="max-w-4xl mx-auto px-6 py-10">

        <h1 class="text-3xl font-black mb-2">Historique des prix</h1>
        <p class="text-slate-400 mb-8">Détection automatique de fraude — SPIKE / DUMP / DUMPING</p>

        <!-- Sélecteur articles -->
        <div class="flex gap-3 flex-wrap mb-8">
            {#each items as item}
                <button
                        onclick={() => selectItem(item)}
                        class="px-4 py-2 rounded-xl text-sm font-medium border transition
            {selectedItem.id === item.id
              ? 'bg-amber-400 text-slate-900 border-amber-400'
              : 'bg-white/5 text-slate-400 border-white/10 hover:border-amber-400/50'}"
                >
                    {item.title}
                </button>
            {/each}
        </div>

        <!-- Prix actuel + variation -->
        <div class="flex items-end gap-4 mb-6">
            {#key lastPrice}
                <div in:fly={{ y: -10, duration: 300 }}>
                    <p class="text-5xl font-black text-white">{lastPrice} €</p>
                </div>
            {/key}
            <p class="text-lg font-bold mb-1 {variation >= 0 ? 'text-green-400' : 'text-red-400'}">
                {variation >= 0 ? '+' : ''}{variation}%
            </p>
        </div>

        <!-- Alerte fraude -->
        {#key alert}
            <div
                    in:fly={{ y: -10, duration: 300 }}
                    class="inline-flex items-center gap-2 border rounded-xl px-4 py-2 mb-8 {alertConfig[alert].bg}"
            >
        <span class="font-semibold text-sm {alertConfig[alert].color}">
          {alertConfig[alert].label}
        </span>
            </div>
        {/key}

        <!-- Graphique SVG -->
        <div class="bg-white/5 border border-white/10 rounded-2xl p-6">
            <svg viewBox="0 0 {W} {H}" class="w-full" preserveAspectRatio="none">
                <defs>
                    <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stop-color="#f59e0b" stop-opacity="0.3"/>
                        <stop offset="100%" stop-color="#f59e0b" stop-opacity="0"/>
                    </linearGradient>
                </defs>

                <!-- Fill sous la courbe -->
                {#if fillPath}
                    <path d={fillPath} fill="url(#grad)" />
                {/if}

                <!-- Ligne principale -->
                {#if svgPath}
                    <path
                            d={svgPath}
                            fill="none"
                            stroke="#f59e0b"
                            stroke-width="2.5"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                    />
                {/if}

                <!-- Point actuel -->
                {#if history.length > 0}
                    {@const prices = history.map(h => h.price)}
                    {@const min = Math.min(...prices)}
                    {@const max = Math.max(...prices)}
                    {@const range = max - min || 1}
                    {@const lx = PAD + ((history.length - 1) / (history.length - 1)) * (W - PAD * 2)}
                    {@const ly = H - PAD - ((prices[prices.length - 1] - min) / range) * (H - PAD * 2)}
                    <circle cx={lx} cy={ly} r="5" fill="#f59e0b" />
                    <circle cx={lx} cy={ly} r="10" fill="#f59e0b" fill-opacity="0.2" />
                {/if}
            </svg>

            <!-- Légende bas -->
            <div class="flex justify-between mt-2 text-xs text-slate-600">
                <span>{history[0]?.time ?? ''}</span>
                <span>{history[history.length - 1]?.time ?? ''}</span>
            </div>
        </div>

    </div>
</div>
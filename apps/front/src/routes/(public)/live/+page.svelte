<script lang="ts">
    import { fly, fade } from 'svelte/transition';
    import { flip } from 'svelte/animate';

    type FeedEvent = {
        id: number;
        type: 'SALE' | 'PRICE_DROP' | 'NEW_ITEM' | 'FRAUD_ALERT';
        title: string;
        body: string;
        time: string;
    };

    const typeConfig = {
        SALE:        { emoji: '💰', color: 'text-green-400',  bg: 'bg-green-400/10  border-green-400/20',  label: 'Vente' },
        PRICE_DROP:  { emoji: '📉', color: 'text-blue-400',   bg: 'bg-blue-400/10   border-blue-400/20',   label: 'Baisse de prix' },
        NEW_ITEM:    { emoji: '✨', color: 'text-amber-400',  bg: 'bg-amber-400/10  border-amber-400/20',  label: 'Nouvel article' },
        FRAUD_ALERT: { emoji: '🚨', color: 'text-red-400',    bg: 'bg-red-400/10    border-red-400/20',    label: 'Alerte fraude' },
    };

    const pool = [
        { type: 'SALE'        as const, title: 'Charizard 1ère éd. vendu',     body: 'Transaction de 350€ finalisée'         },
        { type: 'PRICE_DROP'  as const, title: 'Game Boy Color -30%',           body: 'Prix passé de 120€ à 84€'              },
        { type: 'NEW_ITEM'    as const, title: 'Vinyle Daft Punk Discovery',    body: 'Mis en vente à 65€ — état parfait'     },
        { type: 'FRAUD_ALERT' as const, title: 'SPIKE détecté',                 body: 'Prix x5 en 2h sur Comics Batman #251' },
        { type: 'SALE'        as const, title: 'Figurine Evangelion Unit-01',   body: 'Transaction de 210€ finalisée'         },
        { type: 'NEW_ITEM'    as const, title: 'Montre Seiko SKX007',           body: 'Mis en vente à 280€ — vintage 1996'   },
        { type: 'PRICE_DROP'  as const, title: 'Carte Magic Black Lotus -15%',  body: 'Prix passé de 800€ à 680€'            },
        { type: 'FRAUD_ALERT' as const, title: 'DUMPING détecté',               body: 'Prix anormalement bas sur 3 articles' },
    ];

    let events = $state<FeedEvent[]>([]);
    let counter = $state(0);
    let liveCount = $state(0);
    let totalVolume = $state(0);

    function addEvent() {
        const source = pool[Math.floor(Math.random() * pool.length)];
        const newEvent: FeedEvent = {
            id: counter++,
            ...source,
            time: new Date().toLocaleTimeString('fr-FR'),
        };

        events = [newEvent, ...events].slice(0, 12);
        liveCount++;

        if (source.type === 'SALE') {
            const amount = parseInt(source.body.match(/\d+/)?.[0] ?? '0');
            totalVolume += amount;
        }
    }

    // Simule le WebSocket — on remplacera par le vrai ws:// plus tard
    $effect(() => {
        addEvent();
        const interval = setInterval(addEvent, 2000);
        return () => clearInterval(interval);
    });
</script>

<div class="min-h-screen bg-slate-950 text-white">

    <!-- Header -->
    <div class="border-b border-white/10 px-6 py-4 flex items-center justify-between">
        <a href="/" class="text-2xl font-black">collector<span class="text-amber-400">.shop</span></a>
        <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-green-400 animate-pulse"></span>
            <span class="text-green-400 text-sm font-medium">Live</span>
        </div>
    </div>

    <div class="max-w-3xl mx-auto px-6 py-10">

        <h1 class="text-3xl font-black mb-2">Activité en direct</h1>
        <p class="text-slate-400 mb-8">Flux temps réel des événements sur la marketplace</p>

        <!-- Stats -->
        <div class="grid grid-cols-3 gap-4 mb-8">
            <div class="bg-white/5 border border-white/10 rounded-2xl p-4 text-center">
                <p class="text-3xl font-black text-amber-400">{liveCount}</p>
                <p class="text-slate-400 text-sm mt-1">Événements</p>
            </div>
            <div class="bg-white/5 border border-white/10 rounded-2xl p-4 text-center">
                <p class="text-3xl font-black text-green-400">{totalVolume}€</p>
                <p class="text-slate-400 text-sm mt-1">Volume vendu</p>
            </div>
            <div class="bg-white/5 border border-white/10 rounded-2xl p-4 text-center">
                <p class="text-3xl font-black text-blue-400">2s</p>
                <p class="text-slate-400 text-sm mt-1">Intervalle</p>
            </div>
        </div>

        <!-- Feed -->
        <div class="flex flex-col gap-3">
            {#each events as event (event.id)}
                <div
                        animate:flip={{ duration: 300 }}
                        in:fly={{ y: -40, duration: 400 }}
                        out:fade={{ duration: 200 }}
                        class="flex items-start gap-4 border rounded-2xl p-4 {typeConfig[event.type].bg}"
                >
                    <span class="text-2xl">{typeConfig[event.type].emoji}</span>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2 mb-1">
              <span class="text-xs font-bold uppercase tracking-wider {typeConfig[event.type].color}">
                {typeConfig[event.type].label}
              </span>
                            <span class="text-slate-600 text-xs">{event.time}</span>
                        </div>
                        <p class="font-semibold text-white">{event.title}</p>
                        <p class="text-slate-400 text-sm">{event.body}</p>
                    </div>
                </div>
            {/each}
        </div>

    </div>
</div>
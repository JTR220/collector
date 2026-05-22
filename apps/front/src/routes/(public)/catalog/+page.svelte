<script lang="ts">
    import type { Item } from '$lib/types';

    // Mock data — on branchera l'API plus tard
    const items: Partial<Item>[] = [
        { id: '1', title: 'Carte Pokémon Charizard 1ère éd.', price: 350, category_id: 'cartes', status: 'active' },
        { id: '2', title: 'Figurine Dragon Ball Z Goku', price: 89, category_id: 'figurines', status: 'active' },
        { id: '3', title: 'Console Game Boy originale', price: 120, category_id: 'consoles', status: 'active' },
        { id: '4', title: 'Vinyle Pink Floyd - The Wall', price: 45, category_id: 'musique', status: 'active' },
        { id: '5', title: 'Comics Amazing Spider-Man #1', price: 200, category_id: 'comics', status: 'active' },
        { id: '6', title: 'Montre Casio vintage 1985', price: 75, category_id: 'montres', status: 'sold' },
    ];

    const categories = ['Tous', 'cartes', 'figurines', 'consoles', 'musique', 'comics', 'montres'];
    let selectedCategory = $state('Tous');
    let search = $state('');

    let filtered = $derived(
        items.filter(item =>
            (selectedCategory === 'Tous' || item.category_id === selectedCategory) &&
            item.title?.toLowerCase().includes(search.toLowerCase())
        )
    );
</script>

<div class="min-h-screen bg-slate-950 text-white">

    <!-- Header -->
    <div class="bg-slate-900 border-b border-white/10 px-6 py-4 flex items-center justify-between">
        <a href="/" class="text-2xl font-black">collector<span class="text-amber-400">.shop</span></a>
        <div class="flex gap-3">
            <a href="/login" class="text-slate-400 hover:text-white transition text-sm">Connexion</a>
            <a href="/register" class="bg-amber-400 text-slate-900 font-bold px-4 py-1.5 rounded-lg text-sm hover:bg-amber-300 transition">S'inscrire</a>
        </div>
    </div>

    <div class="max-w-6xl mx-auto px-6 py-10">

        <!-- Search -->
        <div class="mb-8">
            <h1 class="text-3xl font-bold mb-6">Catalogue</h1>
            <input
                    type="text"
                    placeholder="Rechercher un article..."
                    bind:value={search}
                    class="w-full bg-white/5 border border-white/10 rounded-xl px-5 py-3 text-white placeholder-slate-500 focus:outline-none focus:border-amber-400 transition"
            />
        </div>

        <!-- Catégories -->
        <div class="flex gap-2 flex-wrap mb-8">
            {#each categories as cat}
                <button
                        onclick={() => selectedCategory = cat}
                        class="px-4 py-1.5 rounded-full text-sm font-medium transition
            {selectedCategory === cat
              ? 'bg-amber-400 text-slate-900'
              : 'bg-white/5 text-slate-400 hover:bg-white/10'}"
                >
                    {cat}
                </button>
            {/each}
        </div>

        <!-- Grille articles -->
        {#if filtered.length === 0}
            <p class="text-slate-500 text-center py-20">Aucun article trouvé.</p>
        {:else}
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                {#each filtered as item}
                    <a href="/items/{item.id}" class="group bg-white/5 border border-white/10 rounded-2xl overflow-hidden hover:border-amber-400/50 transition">
                        <!-- Placeholder image -->
                        <div class="aspect-square bg-slate-800 flex items-center justify-center text-5xl">
                            🏷️
                        </div>
                        <div class="p-4">
                            <p class="font-semibold text-white group-hover:text-amber-400 transition line-clamp-2">{item.title}</p>
                            <div class="flex items-center justify-between mt-3">
                                <span class="text-xl font-black text-amber-400">{item.price} €</span>
                                {#if item.status === 'sold'}
                                    <span class="text-xs bg-red-500/20 text-red-400 px-2 py-1 rounded-full">Vendu</span>
                                {:else}
                                    <span class="text-xs bg-green-500/20 text-green-400 px-2 py-1 rounded-full">Disponible</span>
                                {/if}
                            </div>
                        </div>
                    </a>
                {/each}
            </div>
        {/if}

    </div>
</div>
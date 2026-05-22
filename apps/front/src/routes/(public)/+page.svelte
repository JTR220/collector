    <script lang="ts">
    import { fly } from 'svelte/transition';

    const pages = [
        {
            href: '/catalog',
            emoji: '🏷️',
            title: 'Catalogue',
            description: 'Parcourez tous les objets de collection disponibles',
            tag: 'Public',
            tagColor: 'bg-slate-700 text-slate-300',
        },
        {
            href: '/live',
            emoji: '📡',
            title: 'Live Feed',
            description: 'Activité en temps réel de la marketplace',
            tag: 'Live',
            tagColor: 'bg-green-400/20 text-green-400',
        },
        {
            href: '/price-tracker',
            emoji: '📈',
            title: 'Price Tracker',
            description: 'Historique des prix et détection de fraude',
            tag: 'Fraude',
            tagColor: 'bg-red-400/20 text-red-400',
        },
        {
            href: '/login',
            emoji: '🔑',
            title: 'Connexion',
            description: 'Accédez à votre espace acheteur ou vendeur',
            tag: 'Auth',
            tagColor: 'bg-amber-400/20 text-amber-400',
        },
    ];
</script>

<div class="min-h-screen bg-slate-950 text-white">

    <!-- Nav -->
    <nav class="border-b border-white/10 px-6 py-4 flex items-center justify-between">
        <span class="text-2xl font-black">collector<span class="text-amber-400">.shop</span></span>
        <div class="flex gap-3 items-center">
            <a href="/login" class="text-slate-400 hover:text-white transition text-sm">Connexion</a>
            <a href="/register" class="bg-amber-400 text-slate-900 font-bold px-4 py-1.5 rounded-lg text-sm hover:bg-amber-300 transition">
                S'inscrire
            </a>
        </div>
    </nav>

    <!-- Hero -->
    <div class="max-w-4xl mx-auto px-6 py-24 text-center">
        <div class="inline-block bg-amber-400/10 text-amber-400 text-sm font-medium px-4 py-1.5 rounded-full mb-6">
            Marketplace de collection entre particuliers
        </div>
        <h1 class="text-6xl font-black leading-tight mb-6">
            Achetez et vendez<br/>vos <span class="text-amber-400">objets rares</span>
        </h1>
        <p class="text-slate-400 text-xl mb-10 max-w-2xl mx-auto">
            Cartes, figurines, vinyles, comics... Trouvez des pépites ou donnez une seconde vie à votre collection.
        </p>
        <div class="flex gap-4 justify-center flex-wrap">
            <a href="/catalog" class="bg-amber-400 text-slate-900 font-bold px-8 py-4 rounded-xl hover:bg-amber-300 transition text-lg">
                Explorer le catalogue
            </a>
            <a href="/login" class="border border-white/20 text-white font-bold px-8 py-4 rounded-xl hover:bg-white/5 transition text-lg">
                Vendre un objet
            </a>
        </div>
    </div>

    <!-- Stats -->
    <div class="border-y border-white/10 py-12 mb-20">
        <div class="max-w-4xl mx-auto px-6 grid grid-cols-3 gap-8 text-center">
            <div>
                <p class="text-4xl font-black text-amber-400">12k+</p>
                <p class="text-slate-400 mt-1">Articles en ligne</p>
            </div>
            <div>
                <p class="text-4xl font-black text-amber-400">3.2k</p>
                <p class="text-slate-400 mt-1">Vendeurs actifs</p>
            </div>
            <div>
                <p class="text-4xl font-black text-amber-400">5%</p>
                <p class="text-slate-400 mt-1">Commission seulement</p>
            </div>
        </div>
    </div>

    <!-- Pages dispo -->
    <div class="max-w-4xl mx-auto px-6 pb-24">
        <h2 class="text-2xl font-bold mb-8 text-slate-300">Explorer</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {#each pages as page, i}

                href={page.href}
                in:fly={{ y: 20, duration: 300, delay: i * 80 }}
                class="group bg-white/5 border border-white/10 rounded-2xl p-6 hover:border-amber-400/40 hover:bg-white/8 transition"
                >
                <div class="flex items-start justify-between mb-4">
                    <span class="text-4xl">{page.emoji}</span>
                    <span class="text-xs font-semibold px-2 py-1 rounded-full {page.tagColor}">
              {page.tag}
            </span>
                </div>
                <h3 class="text-lg font-bold text-white group-hover:text-amber-400 transition mb-1">
                    {page.title}
                </h3>
                <p class="text-slate-400 text-sm">{page.description}</p>
                </a>
            {/each}
        </div>
    </div>

</div>
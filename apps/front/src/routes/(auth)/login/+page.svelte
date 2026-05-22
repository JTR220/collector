<script lang="ts">
    let email = $state('');
    let password = $state('');

    function handleSubmit() {
        console.log('email:', email, 'password:', password);
    }
</script>

<div class="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 flex items-center justify-center p-4">
    <div class="w-full max-w-md">

        <!-- Logo / Brand -->
        <div class="text-center mb-8">
            <h1 class="text-4xl font-black text-white tracking-tight">
                collector<span class="text-amber-400">.shop</span>
            </h1>
            <p class="text-slate-400 mt-2 text-sm">Connectez-vous à votre compte</p>
        </div>

        <!-- Card -->
        <div class="bg-white/5 backdrop-blur border border-white/10 rounded-2xl p-8">

            <div class="flex flex-col gap-5">
                <div class="flex flex-col gap-1.5">
                    <label class="text-slate-300 text-sm font-medium">Email</label>
                    <input
                            type="email"
                            placeholder="vous@example.com"
                            bind:value={email}
                            class="bg-white/10 border border-white/20 text-white placeholder-slate-500 rounded-xl px-4 py-3 focus:outline-none focus:border-amber-400 transition"
                    />
                </div>

                <div class="flex flex-col gap-1.5">
                    <label class="text-slate-300 text-sm font-medium">Mot de passe</label>
                    <input
                            type="password"
                            placeholder="••••••••"
                            bind:value={password}
                            class="bg-white/10 border border-white/20 text-white placeholder-slate-500 rounded-xl px-4 py-3 focus:outline-none focus:border-amber-400 transition"
                    />
                </div>

                <button
                        onclick={handleSubmit}
                        class="bg-amber-400 hover:bg-amber-300 text-slate-900 font-bold py-3 rounded-xl transition mt-2"
                >
                    Se connecter
                </button>
            </div>

            <p class="text-center text-slate-400 text-sm mt-6">
                Pas encore de compte ?
                <a href="/register" class="text-amber-400 hover:underline font-medium">S'inscrire</a>
            </p>
        </div>

    </div>
</div>
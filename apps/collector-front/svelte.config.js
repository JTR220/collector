import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// adapter-node : build de production autonome (node build) pour l'image
		// Docker et le déploiement Kubernetes. `vite dev` n'est pas impacté.
		adapter: adapter()
	}
};

export default config;

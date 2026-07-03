# Étape 5 — image de PRODUCTION : build multi-stage SvelteKit (adapter-node),
# image finale minimale en non-root. Pour le dev local via docker-compose,
# utiliser Dockerfile.dev (serveur Vite avec HMR).
#
# Les URLs d'API sont lues AU RUNTIME via $env/dynamic/public (PUBLIC_*) :
# l'image est agnostique de l'environnement — une seule image pour staging et
# prod, configurée par variables de pod (ConfigMap collector-front-config).
FROM node:26-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

RUN npm run build && npm prune --omit=dev

FROM node:26-alpine

WORKDIR /app

# Durcissement : applique les patchs Alpine du jour et retire les outils de
# build (npm, corepack, yarn) inutiles au runtime — `node build` suffit.
# hadolint ignore=DL3017
RUN apk upgrade --no-cache \
    && rm -rf /usr/local/lib/node_modules /usr/local/bin/npm /usr/local/bin/npx \
       /usr/local/bin/corepack /opt/yarn* /usr/local/bin/yarn /usr/local/bin/yarnpkg

ENV NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=3000

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

USER node

CMD ["node", "build"]

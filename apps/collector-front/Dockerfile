# Étape 5 — image de PRODUCTION : build multi-stage SvelteKit (adapter-node),
# image finale minimale en non-root. Pour le dev local via docker-compose,
# utiliser Dockerfile.dev (serveur Vite avec HMR).
#
# Les variables VITE_* sont compilées dans le bundle AU BUILD : cette image
# est donc liée à un environnement (build args injectés par la CI).
FROM node:26-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_AUTH_API_BASE_URL=http://localhost:8080
ARG VITE_CATALOG_API_BASE_URL=http://localhost:8081
ARG VITE_PRICE_TRACKER_API_BASE_URL=http://localhost:8082
ARG VITE_NOTIFICATION_API_BASE_URL=http://localhost:8083
ENV VITE_AUTH_API_BASE_URL=${VITE_AUTH_API_BASE_URL} \
    VITE_CATALOG_API_BASE_URL=${VITE_CATALOG_API_BASE_URL} \
    VITE_PRICE_TRACKER_API_BASE_URL=${VITE_PRICE_TRACKER_API_BASE_URL} \
    VITE_NOTIFICATION_API_BASE_URL=${VITE_NOTIFICATION_API_BASE_URL}

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

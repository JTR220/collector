# Pipeline DevSecOps GitOps — collector.shop

Implémentation des 15 étapes du cours « Pipeline DevSecOps GitOps avec
Argo CD ». Principe directeur : **chaîne de confiance** — les tests
garantissent que le code fonctionne, les scans qu'il est sain, la signature
qu'il n'a pas été altéré, Argo CD qu'il est déployé fidèlement, Kyverno que
rien ne contourne le tout, et l'observabilité qu'on le voit.

## Mapping des étapes

| # | Étape | Outil | Implémentation | Statut |
|---|-------|-------|----------------|--------|
| 1 | Tests & lint | `go test -race`, `go vet`, golangci-lint, vitest, eslint+prettier | [backend.yml](../.github/workflows/backend.yml) `lint-test` · [frontend.yml](../.github/workflows/frontend.yml) `quality`/`tests` | ✅ bloquant |
| 1bis | Tests e2e | Playwright (auth-service + catalog-service réels via docker compose, pas de mock) | [frontend.yml](../.github/workflows/frontend.yml) `e2e` · specs [apps/collector-front/e2e/](../apps/collector-front/e2e/) | ✅ bloquant — 2e type de test du pipeline |
| 2 | SAST | gosec (SARIF), Semgrep, SonarCloud | backend `sast`+`sonarcloud` · frontend `sast` | ✅ bloquant |
| 3 | Secrets scanning | gitleaks (historique complet) | [gitleaks.yml](../.github/workflows/gitleaks.yml) — tous push/PR, tous chemins | ✅ bloquant |
| 4 | SCA dépendances | govulncheck, Trivy fs, npm audit, Dependabot | backend `sca` · frontend `sca` · [dependabot.yml](../.github/dependabot.yml) | ✅ bloquant |
| 5 | Image multi-stage | distroless static **nonroot**, hadolint | Dockerfiles des 4 services Go + front adapter-node ([Dockerfile](../apps/collector-front/Dockerfile)) | ✅ |
| 6 | Scan d'image | Trivy, exit-code 1 sur CRITICAL/HIGH, **avant** push | jobs `image` (build local → scan → push) | ✅ bloquant |
| 7 | SBOM | Syft (SPDX json), attaché à l'image | jobs `image` — `anchore/sbom-action` + `cosign attest` | ✅ |
| 8 | Signature | cosign keyless (OIDC GitHub → Rekor) | jobs `image` — `cosign sign` sur le digest | ✅ |
| 9 | Registre | GHCR, tags par SHA, digest dans les manifests | `docker/metadata-action` (`sha-…`), jamais `latest` | ✅ |
| 10 | Manifests + Kustomize | base + overlays staging/prod | [infra/k8s/](../infra/k8s/) — voir [infra/README.md](../infra/README.md) | ✅ |
| 11 | Argo CD | App of Apps, sync auto staging / manuel prod, selfHeal+prune | [infra/argocd/](../infra/argocd/) | 🔶 cluster bootstrappé (VPS), staging en cours de convergence |
| 12 | Secrets cluster | Sealed Secrets (kubeseal) | [infra/secrets/](../infra/secrets/) + app `sealed-secrets` | ✅ générés pour staging et prod |
| 13 | Rollout progressif | Argo Rollouts, canary 25 %/pause/100 % | [infra/k8s/addons/rollouts/](../infra/k8s/addons/rollouts/) | 🔶 optionnel, prêt à brancher |
| 14 | Policies runtime | Kyverno : verifyImages (cosign), no-latest, non-root, resources | [infra/policies/](../infra/policies/) | ✅ **Enforce** (basculé le 10/07/2026, à revalider sur `kubectl get policyreport -A`) |
| 15 | Observabilité | kube-prometheus-stack + métriques applicatives | app `monitoring` + [catalog-service/metrics](../apps/backend/catalog-service/metrics/metrics.go) | ✅ stack + `catalog-service` instrumenté (autres services à faire) |

## Flux d'un push sur main

```
push apps/backend/** ─┬─► lint-test (gofmt, vet, golangci-lint, go test -race)
                      ├─► sast (gosec → SARIF)            } parallèles,
                      ├─► sca (govulncheck)               } tous bloquants
                      └─► gitleaks (workflow dédié)
                                 │ tous verts
                                 ▼
                      image : hadolint → build multi-stage → Trivy (bloquant)
                              → push GHCR sha-<commit> → SBOM syft
                              → cosign sign + attest (keyless)
                                 │
                                 ▼
                      update-manifests : kustomize edit set image <digest>
                              → commit dans infra/k8s/overlays/staging/
                                 │
                                 ▼
                      Argo CD : sync auto staging (Kyverno vérifie la signature
                              à l'admission). Prod = promotion par PR + Sync manuel.
```

Sur **pull request** : tout tourne (y compris build + scan d'image) mais rien
n'est poussé ni signé ni déployé.

## Garde-fous notables

- **La CI n'a aucun credential cluster** : elle ne fait qu'un commit Git
  (`update-manifests`, `permissions: contents: write`). Un push de
  `github-actions[bot]` ne redéclenche pas de workflow (anti-boucle GitHub) et
  les deux workflows partagent le groupe de concurrence
  `gitops-update-manifests` (pas de push concurrents).
- **Build once** : l'image est buildée puis scannée **localement** ; seule une
  image saine est poussée, et c'est son **digest** (immuable) qui est déployé.
- **Sonar** : un seul scan pour tout le backend — deux scans matrix avec le
  même `projectKey` s'écrasaient mutuellement dans SonarCloud.

## Limites connues / prochains chantiers

1. **Front : URLs d'API compilées au build** (`import.meta.env.VITE_*`).
   L'image est buildée avec les hosts *staging* ; une vraie prod front demande
   soit un build dédié, soit (mieux) un refactor vers `$env/dynamic/public`
   pour lire les URLs au runtime → restaure le « build once, deploy many ».
2. **Couverture de tests backend inégale** : de nombreux `*_test.go` couvrent
   désormais controllers/middlewares/repository/PII/modération sur les 4
   services, mais la couverture n'est pas homogène partout — à surveiller via
   l'indicateur de couverture CI plutôt qu'à considérer comme un trou uniforme.
3. **Instrumentation Prometheus des services Go** : `catalog-service` expose
   `/metrics` (requêtes HTTP, articles, commandes, modération, uploads) ;
   `auth-service`, `notification-service` et `price-tracker-service` restent
   à instrumenter pour une observabilité homogène et pour les ServiceMonitors/
   analyse canary (étape 13).
4. **Kyverno passé en Enforce** (10/07/2026) : à revalider contre
   `kubectl get policyreport -A` sur le cluster réel avant le prochain sync
   Argo CD, pour s'assurer qu'aucun pod existant n'est bloqué à tort.
5. **notification-service / price-tracker-service** : Dockerfiles durcis mais
   pas encore dans la CI ni les manifests (services non intégrés).
6. **Dette eslint front** : 3 règles svelte désactivées et `no-unused-vars` en
   warning (110 occurrences préexistantes) pour permettre un lint bloquant —
   voir le commentaire dans
   [eslint.config.js](../apps/collector-front/eslint.config.js), à réactiver
   au fil des refactors.

## État du bootstrap cluster (06/07/2026)

Cluster réel déployé sur un **VPS OVH** (`vps-a6745acc.vps.ovh.net`, VPS-2 2027 —
4 vCore/8 Go RAM/75 Go SSD, Gravelines), pas sur k3d local. Adaptation du
bootstrap [infra/README.md](../infra/README.md) : `curl -sfL https://get.k3s.io | sh -`
à la place de `k3d cluster create`.

Progression :

- [x] k3s installé, node `Ready`
- [x] kubectl + Helm configurés pour l'utilisateur non-root
- [x] Argo CD installé (Helm), tous les pods `1/1 Running`
- [x] `kubectl apply -f infra/argocd/bootstrap/root-app.yaml` → `collector-root` `Synced`/`Healthy`
- [x] `argo-rollouts`, `kyverno-policies` → `Synced`/`Healthy`
- [x] `sealed-secrets` → `Synced`/`Healthy` (l'URL du repo Helm avait bougé,
      `bitnami-labs.github.io` → `bitnami.github.io` : voir
      [infra/argocd/apps/sealed-secrets.yaml](../infra/argocd/apps/sealed-secrets.yaml))
- [x] `collector-staging` → **`Synced`/`Healthy`**, les 7 pods `1/1 Running` :
      SealedSecret `collector-secrets` généré via
      [infra/secrets/README.md](../infra/secrets/README.md), et deux bugs corrigés
      au passage (probe RabbitMQ `timeout` → `timeoutSeconds` dans
      `infra/k8s/base/rabbitmq/statefulset.yaml` ; variables d'exchange/queue
      RabbitMQ manquantes dans les manifests de `price-tracker-service` et
      `notification-service`, présentes seulement dans `docker-compose.yml`)
- [ ] `kyverno`, `monitoring` → à reconfirmer `Healthy` (étaient encore
      `OutOfSync`/`Progressing` lors du bootstrap initial)
- `collector-prod` reste `OutOfSync`/`Missing` : **attendu**, sync manuel volontaire (pas une erreur)

Le flux complet (push → CI → digest → Argo CD sync auto) est validé de bout en
bout sur staging. SealedSecrets générés pour staging et prod, Kyverno passé en
Enforce (10/07/2026). Reste à faire : sauvegarde de la clé privée du
controller sealed-secrets hors cluster (voir `infra/secrets/README.md`).

## Actions manuelles côté GitHub (une fois)

- [ ] `SONAR_TOKEN` déjà en place (réutilisé).
- [ ] Activer **Secret Scanning + Push Protection** (Settings → Code security).
- [ ] **Protection de branche `main`** : PR + checks obligatoires (gitleaks,
      lint-test, sast, sca, image) — sinon les jobs « bloquants » restent des
      conventions contournables.
- [ ] Vérifier la visibilité des packages GHCR après le premier push
      (Settings du package → public, ou créer un `imagePullSecret` en lecture
      seule pour le cluster — jamais le token d'écriture).
- [ ] Bootstrap du cluster : voir [infra/README.md](../infra/README.md).

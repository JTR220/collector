# Infra GitOps — collector.shop

L'état désiré du cluster est décrit ici ; **Argo CD** fait converger la réalité
vers cette description en continu. Tout déploiement est un commit (auditable),
tout rollback est un revert (trivial), et la CI n'a **jamais** les credentials
du cluster.

> Le cours recommande un repo `collector-infra` séparé ; ici l'infra vit dans
> `infra/` du monorepo — même garanties GitOps, un seul repo à gérer. Les
> workflows CI sont filtrés par chemins : un commit dans `infra/` ne
> redéclenche pas de build. Le dossier est extractible tel quel vers un repo
> dédié plus tard.

## Arborescence

```
infra/
├── argocd/
│   ├── bootstrap/root-app.yaml   # App of Apps : LE manifest à appliquer à la main
│   └── apps/                     # 1 fichier = 1 application gérée par Argo CD
├── k8s/
│   ├── base/                     # description unique des services (Kustomize)
│   ├── overlays/
│   │   ├── staging/              # sync auto — digests épinglés par la CI
│   │   └── prod/                 # sync MANUEL — promotion par PR
│   └── addons/rollouts/          # canary Argo Rollouts (optionnel)
├── policies/                     # ClusterPolicies Kyverno
└── secrets/                      # templates + mode d'emploi Sealed Secrets
```

## Flux complet

```
push code ──► CI GitHub Actions (tests, scans, build, signature)
                      │
                      ▼  commit du digest dans overlays/staging/
              repo Git (source de vérité)
                      │
                      ▼  pull (jamais de push vers le cluster)
              Argo CD ──► cluster (staging auto, prod sur clic Sync)
```

## Bootstrap (Phase 1 du cours — cluster local k3d)

```bash
# 0. Pré-requis : docker, k3d, kubectl, helm, kubeseal

# 1. Cluster local avec port 80 exposé (Traefik ingress inclus dans k3s)
k3d cluster create collector --port "80:80@loadbalancer"

# 2. Argo CD
kubectl create namespace argocd
helm repo add argo https://argoproj.github.io/argo-helm
helm install argocd argo/argo-cd -n argocd

# 3. Mot de passe initial + UI
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
kubectl port-forward svc/argocd-server -n argocd 8443:443   # → https://localhost:8443 (admin)

# 4. LA commande GitOps : tout le reste découle de ce manifest
kubectl apply -f infra/argocd/bootstrap/root-app.yaml
```

Argo CD installe alors lui-même : sealed-secrets, kyverno (+ policies),
monitoring, argo-rollouts, collector-staging et collector-prod.

```bash
# 5. Secrets (les apps restent Degraded tant que collector-secrets n'existe pas)
#    → suivre infra/secrets/README.md (kubeseal), committer les SealedSecrets

# 6. Résolution des hosts en local (fichier hosts, en admin) :
# 127.0.0.1 collector.staging.local auth.collector.staging.local api.collector.staging.local
# 127.0.0.1 collector.local auth.collector.local api.collector.local
```

Premier déploiement : pousser un commit sur `apps/backend/` ou
`apps/collector-front/` → la CI build, scanne, signe, pousse l'image et commit
le digest dans `overlays/staging/` → Argo CD synchronise. Zéro action manuelle.

## Promotion staging → prod

1. Ouvrir `infra/k8s/overlays/staging/kustomization.yaml`, copier le digest
   validé (`digest: sha256:…`).
2. Le coller dans `infra/k8s/overlays/prod/kustomization.yaml` **via une PR**
   (protection de branche recommandée).
3. Merger, puis cliquer **Sync** sur `collector-prod` dans l'UI Argo CD — ce
   clic est la validation de mise en production, avec diff visible avant.

## Webhook (optionnel, sync instantané)

Par défaut Argo CD poll le repo toutes les 3 min. Pour un sync immédiat :
GitHub → Settings → Webhooks → `https://<argocd>/api/webhook` (content type
`application/json`).

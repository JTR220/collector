# Secrets — génération des SealedSecrets (staging & prod)

Les services lisent leurs secrets depuis un `Secret` Kubernetes nommé
**`collector-secrets`** (clés : `JWT_SECRET`, `DB_PASSWORD`, `RABBITMQ_URL`,
`RABBITMQ_DEFAULT_USER/PASS`, `PRICE_DATABASE_DSN`, `NOTIF_DATABASE_DSN`).

On ne committe **jamais** ce Secret en clair. On committe un **SealedSecret**
chiffré avec la clé publique du cluster : seul le controller `sealed-secrets`
(déployé via [`infra/argocd/apps/sealed-secrets.yaml`](../argocd/apps/sealed-secrets.yaml))
peut le déchiffrer, in-cluster. Le fichier chiffré est donc commitable sans risque.

## Prérequis
- `kubeseal` installé (même version majeure que le controller).
- Accès `kubectl` au cluster, controller `sealed-secrets-controller` running
  dans `kube-system` (l'app Argo CD `sealed-secrets` doit être synchronisée).

## Procédure (à répéter pour chaque environnement)

Faire une passe pour `collector-staging`, puis une pour `collector-prod`.

### 1. Copier le template et remplir les vraies valeurs
```bash
cd infra/secrets
cp collector-secrets.template.yaml collector-secrets.unsealed.yaml   # *.unsealed.yaml est gitignoré
```
Éditer `collector-secrets.unsealed.yaml` :
- `metadata.namespace` : `collector-staging` (ou `collector-prod`).
- Générer chaque valeur : `openssl rand -base64 32`.
- ⚠️ `JWT_SECRET` doit être **identique** pour auth-service, catalog-service et
  notification-service (c'est le même Secret, donc automatique).
- Reporter le mot de passe DB choisi dans `DB_PASSWORD` **et** dans les DSN
  (`PRICE_DATABASE_DSN`, `NOTIF_DATABASE_DSN`) et l'URL (`RABBITMQ_URL`).

### 2. Sceller avec kubeseal
```bash
kubeseal \
  --controller-name sealed-secrets-controller \
  --controller-namespace kube-system \
  --format yaml \
  < collector-secrets.unsealed.yaml \
  > sealed-collector-secrets.yaml
```
Puis placer le fichier produit dans l'overlay correspondant :
```bash
mv sealed-collector-secrets.yaml ../k8s/overlays/staging/   # ou overlays/prod/
rm collector-secrets.unsealed.yaml                          # ne jamais committer le clair
```

### 3. Activer la ressource dans l'overlay
Dé-commenter la ligne dans `infra/k8s/overlays/<env>/kustomization.yaml` :
```yaml
resources:
  - ../../base
  - ingress.yaml
  - sealed-collector-secrets.yaml   # <-- dé-commenter
```

### 4. Valider et committer
```bash
kubectl kustomize infra/k8s/overlays/staging   # doit builder sans erreur
git add infra/k8s/overlays/staging/sealed-collector-secrets.yaml \
        infra/k8s/overlays/staging/kustomization.yaml
git commit -m "chore(secrets): SealedSecret collector-secrets pour staging"
```
Argo CD (staging = auto-sync) applique le SealedSecret ; le controller le
déchiffre en `Secret collector-secrets`. Pour la prod, le Sync reste manuel.

## À ne jamais faire
- Committer `*.unsealed.yaml` ou un `Secret` en clair (le `.gitignore` couvre
  `*.unsealed.yaml`, mais reste vigilant).
- Réutiliser les mêmes valeurs entre staging et prod : génère des secrets
  distincts par environnement.

## Rotation / restauration
Sauvegarder la clé privée du controller (sinon les SealedSecrets existants
deviennent indéchiffrables après réinstallation) :
```bash
kubectl get secret -n kube-system \
  -l sealedsecrets.bitnami.com/sealed-secrets-key -o yaml \
  > sealed-secrets-key.backup.yaml   # à stocker hors dépôt, en lieu sûr
```

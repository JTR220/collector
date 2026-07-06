# Étape 13 — Déploiement progressif (Argo Rollouts)

Exemple **prêt à l'emploi mais non branché** : il nécessite que le controller
Argo Rollouts soit installé (application `argo-rollouts` dans
`infra/argocd/apps/`) avant d'être ajouté à un overlay.

## Principe

`rollout-catalog-service.yaml` utilise `workloadRef` : le Rollout référence le
Deployment existant de catalog-service sans le dupliquer. À chaque nouvelle
image, le trafic bascule progressivement :

1. **25 %** du trafic vers la nouvelle version ;
2. **pause** — vérification manuelle (logs, dashboards Grafana), puis :
   ```bash
   kubectl argo rollouts promote catalog-service -n collector-prod
   ```
3. **100 %** — promotion complète.

Une régression ne touche donc jamais que 25 % des requêtes, et le rollback est
instantané (`kubectl argo rollouts undo`).

## Activation

1. Vérifier que l'app `argo-rollouts` est synchronisée dans Argo CD.
2. Ajouter dans `infra/k8s/overlays/prod/kustomization.yaml` :
   ```yaml
   resources:
     - ../../addons/rollouts/rollout-catalog-service.yaml
   ```
3. Installer le plugin kubectl :
   ```bash
   # https://argoproj.github.io/argo-rollouts/installation/#kubectl-plugin-installation
   kubectl argo rollouts dashboard   # UI locale sur :3100
   ```

## Étape suivante : analyse automatique

Quand les services Go exposeront `/metrics` (prometheus/client_golang), la
pause manuelle peut être remplacée par un `AnalysisTemplate` branché sur
Prometheus (taux d'erreur, latence p99) : promotion automatique si les
métriques sont saines, rollback automatique sinon. Squelette commenté dans
`rollout-catalog-service.yaml`.

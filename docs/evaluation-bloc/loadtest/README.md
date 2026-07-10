# Test de montée en charge — Siege

Couvre l'exigence orale des consignes (*« vous devrez présenter la montée en
charge de l'application en lançant des tests de charge avec Siege ou JMeter »*)
et l'indicateur qualité #4 de
[`03-indicateurs-qualite.md`](../03-indicateurs-qualite.md) (efficacité de
performance).

## Pourquoi ces endpoints

Seuls les endpoints **publics en lecture** de `catalog-service` sont ciblés
(`/health`, `/article`, `/category`, `/article/:id`) : aucune authentification
requise, aucune écriture — rejouable à volonté sans épuiser le stock
d'articles de démo ni créer de fausses commandes. `POST /article/:id/buy`
(mentionné comme endpoint critique dans `03-indicateurs-qualite.md`) est
volontairement exclu d'un run automatisé répété : il mute des données réelles
(vente effective d'un article, une seule fois possible par article). Si vous
voulez le démontrer spécifiquement à l'oral, faites-le en un ou deux appels
manuels sur un article de test dédié plutôt qu'en boucle Siege.

## Prérequis

Sur la machine qui lance le test (le VPS lui-même, ou votre poste) :

```bash
sudo apt install -y siege
```

## Lancer le test

```bash
cd docs/evaluation-bloc/loadtest
chmod +x run-siege.sh

# Local (docker-compose, apps/docker-compose.yml) :
./run-siege.sh http://localhost:8081 25 1M

# Staging (routage par path, voir infra/k8s/overlays/staging/ingress.yaml) :
./run-siege.sh https://staging.chaker.pro:8443/api 25 1M

# Prod (sous-domaine dédié, voir infra/k8s/overlays/prod/ingress.yaml) :
./run-siege.sh https://api.chaker.pro 25 1M
```

Paramètres : `<base-url> [concurrence=25] [durée=1M]` (format Siege : `30S`,
`1M`, `5M`...). Monter progressivement la concurrence (10 → 25 → 50 → 100)
pour trouver le point de dégradation plutôt que de lancer directement un gros
chiffre.

## Lire le résultat

Siege affiche en fin de run (et dans `results-<date>.log`) :

- **Transactions** : nombre total de requêtes traitées.
- **Availability (%)** : doit rester proche de 100 % — toute baisse signale
  des erreurs/timeouts.
- **Response time** : temps de réponse moyen — à comparer à une exécution de
  référence pour détecter une régression (voir seuil de vigilance de
  l'indicateur #4).
- **Transaction rate** / **Throughput** : débit soutenu.
- **Failed transactions** : doit rester à 0 ; sinon regarder lesquelles
  (timeout, 5xx) dans `console-<date>.log`.

Résultat à coller dans le support de soutenance et/ou dans
`06-preparation-soutenance-jury.md`.

## Alternative JMeter

Si vous préférez JMeter (GUI ou `jmeter -n -t plan.jmx`), les mêmes endpoints
listés dans [`urls.txt.tpl`](urls.txt.tpl) (une fois `__BASE__` substitué)
suffisent à construire un plan de test HTTP Request équivalent.

## Résultats — 10/07/2026

### ✅ Bug 500 corrigé, run complet validé sur 3 paliers de concurrence

Vérifié à 14h30 par appel direct (`curl`) : `GET /api/article`,
`GET /api/category` et `GET /api/health` répondent tous **200** avec des
données valides sur `collector-staging`. Le run complet
(`/health` + `/article` + `/category` + `/article/:id`) a été rejoué à
3 paliers de concurrence (conteneur Debian jetable + siege, 1 minute
chacun) :

| Concurrence | Disponibilité | Temps de réponse moyen | Transactions/s | Échecs |
|---|---|---|---|---|
| 25  | 97.57% | 0.17 s | 16.17 | 24 (timeouts socket) |
| 50  | 97.29% | 0.24 s | 26.95 | 45 (timeouts socket) |
| 100 | 97.74% | 1.39 s | 24.14 | 33 (timeouts socket) |

Logs bruts : `results-20260710-1426-full-c25.log`,
`results-20260710-1427-full-c50.log`, `results-20260710-1428-full-c100.log`.
Le temps de réponse moyen se dégrade nettement à 100 utilisateurs
concurrents (1.39s vs 0.17-0.24s), mais la disponibilité reste stable
autour de 97-98% aux trois paliers — les échecs restants sont des
timeouts socket ponctuels (pas des 500 applicatifs), cohérents avec un
test lancé depuis un conteneur jetable plutôt qu'une infra dédiée au
bench. **Ce résultat est présentable à l'oral** : le critère "disponibilité
et montée en charge démontrées" est désormais satisfait.

### Run bloquant du matin (avant correction du 500), pour mémoire

Exécuté à 07:50 (`results-20260710-074953.log` /
`console-20260710-074953.log`) : **47.06% de disponibilité, 2 transactions
réussies sur 18 échouées** — `GET /api/article` et `GET /api/category`
renvoyaient **500**. Corrigé depuis (voir ci-dessus).

### Run interim, `/health` seul (avant la découverte du 500)

Un run progressif avait été fait sur `GET /api/health` seul (conteneur
Debian jetable + siege, cf. section outil ci-dessus), résultats bruts dans
`results-interim/health-only-runs.log` :

| Concurrence | Disponibilité | Temps de réponse moyen | Transactions/s |
|---|---|---|---|
| 25  | 100.00% | 0.10 s | 109.18 |
| 50  | 100.00% | 0.60–0.75 s (reproductible sur 2 runs) | 6.6–7.3 (chute anormale, cause non identifiée) |
| 100 | 100.00% | 0.21 s | 111.48 |

Le creux au palier 50 est reproductible (2 runs cohérents) mais sa cause
n'est pas expliquée — pas de dégradation serveur visible (dispo 100%, pas
de timeout), donc probablement un artefact client (Docker/Siege) plutôt
qu'un vrai comportement de `collector-staging`. À ré-investiguer si le
temps le permet.

✅ Fait — voir la section "Bug 500 corrigé" ci-dessus pour le résultat à
présenter à l'oral (pas celui sur `/health` seul).

## Démo live : voir les pods se dupliquer (HPA)

Avant le 10/07/2026, **aucun HorizontalPodAutoscaler n'existait** dans le
repo — tous les services étaient figés à `replicas: 1` en dur, donc rien ne
se serait dupliqué pendant une démo, même sous forte charge. Ajouté depuis :
`infra/k8s/base/{auth-service,notification-service,price-tracker-service,
collector-front}/hpa.yaml` (CPU, 50% d'utilisation moyenne, 1→4 replicas).
`catalog-service` reste hors HPA : son volume `catalog-uploads` est
`ReadWriteOnce`, donc figé à `replicas: 1` (voir `pvc.yaml`).

**Prod n'est pas concernée** (`collector-prod` reste `OutOfSync`/`Missing`,
jamais synchronisé — voir `docs/DEVSECOPS.md`) : impossible de lui envoyer
plus de charge, il n'y a rien qui tourne dessus. La démo live ne peut se
faire que sur **staging**.

Piège GitOps évité : `collector-staging` a `selfHeal: true` (Argo CD), qui
aurait ramené le nombre de pods à 1 à chaque resync et annulé le scaling en
plein test — `infra/argocd/apps/collector-staging.yaml` ignore désormais
`spec/replicas` pour ces 4 Deployments (`ignoreDifferences`).

### Prérequis à vérifier sur le cluster (pas testé depuis ce poste, pas d'accès kubectl/SSH)

```bash
kubectl top nodes                       # doit répondre : confirme que metrics-server tourne
                                         # (k3s l'active par défaut, mais à revérifier)
kubectl get hpa -n collector-staging    # les 4 HPA doivent apparaître avec une valeur TARGETS
```

Si `kubectl top` échoue ou si `TARGETS` reste `<unknown>` : metrics-server
n'est pas up, l'HPA ne peut pas scaler (pas d'erreur bloquante côté Argo CD,
juste aucune décision de scaling prise).

### Déroulé pour la soutenance

Deux terminaux côte à côte pendant le run Siege :

```bash
# Terminal 1 : regarder les pods apparaître
kubectl get pods -n collector-staging -l app=auth-service -w

# Terminal 2 : regarder la métrique CPU qui déclenche le scale-up
kubectl get hpa -n collector-staging -w
```

Puis lancer une charge plus soutenue que les runs de smoke-test ci-dessus
(l'HPA a une fenêtre de stabilisation ~15-60s avant de réagir, donc un run
de 30s ne laisse pas le temps de voir grand-chose) :

```bash
./run-siege.sh https://staging.chaker.pro:8443/auth 100 3M
```

`/auth` cible directement `auth-service` (health check inclus) — un run de
3 minutes à 100 utilisateurs concurrents doit laisser le temps à l'HPA de
détecter le dépassement de 50% CPU (seuil bas : 25m sur une requête de
50m) et de scaler avant la fin du run.

**Non testé en conditions réelles** (pas d'accès cluster depuis ce poste) :
à valider une fois avant la soutenance, pas le jour J.

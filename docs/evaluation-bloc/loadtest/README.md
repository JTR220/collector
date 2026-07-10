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

## Résultats — 10/07/2026 (INTERIM, `/health` seul)

⚠️ **Bug bloquant découvert en lançant le run complet** :
`GET /api/article` et `GET /api/category` renvoient **500**
(`{"error":"Impossible de recuperer les articles"}` /
`{"error":"Impossible de recuperer les categories"}`) sur
`collector-staging`, alors que `GET /api/health` répond `200`. Suspect :
décalage entre le code déployé (filtre de modération `pending_review` sur
`GetAllArticles`/`GetAllCategories`, voir
[articleController.go](../../../apps/backend/catalog-service/controllers/articleController.go))
et le schéma/les données réelles de la base staging — la table `articles`
n'a peut-être pas la colonne `status` attendue, ou Argo CD n'a pas encore
synchronisé la dernière image. Pas d'accès `kubectl`/SSH direct au cluster
depuis ce poste pour confirmer la cause exacte.

**En attendant la correction**, un run progressif a été fait sur
`GET /api/health` seul (conteneur Debian jetable + siege, cf. section
outil ci-dessus), résultats bruts dans `results-interim/health-only-runs.log` :

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

**À faire avant la soutenance** : corriger le 500 sur `/article`/`/category`,
puis relancer `./run-siege.sh https://staging.chaker.pro:8443/api 25 1M`
(et paliers 50/100) pour obtenir un vrai résultat de montée en charge sur
le catalogue métier — c'est ce résultat-là, pas celui sur `/health` seul,
qui doit être présenté à l'oral.

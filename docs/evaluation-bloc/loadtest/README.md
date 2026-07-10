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

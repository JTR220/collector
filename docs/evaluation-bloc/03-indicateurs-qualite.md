# Indicateurs qualité logicielle (ISO 25010)

Quatre indicateurs, chacun mesuré automatiquement par un outil déjà intégré
au pipeline CI (`.github/workflows/backend.yml` / `frontend.yml`), choisis
pour couvrir des exigences ISO 25010 distinctes sans chercher l'exhaustivité.

## 1. Fiabilité (*Reliability*) — couverture des tests automatisés

- **Mesure** : `go test -race -covermode=atomic -coverprofile=coverage.out`
  sur les 4 services Go, rapport remonté à SonarCloud
  (`sonar.go.coverage.reportPaths`) ; côté front, vitest sur les clients API
  et stores critiques.
- **Seuil de vigilance** : toute baisse de couverture sur un module
  transactionnel (`marketplaceController`, `authclient`, middlewares JWT) doit
  déclencher une revue — ces chemins concentrent le risque financier et
  sécurité de l'application.
- **Lien dette technique** : un code non couvert est un code qui peut
  régresser silencieusement à chaque modification ultérieure ; sans ce
  filet, chaque nouvelle fonctionnalité augmente le risque de régression non
  détectée avant la mise en production — la dette s'accumule de façon
  invisible jusqu'à l'incident.

## 2. Sécurité (*Security*) — vulnérabilités bloquantes détectées

- **Mesure** : nombre de findings CRITICAL/HIGH remontés par **Trivy**
  (image + filesystem), **gosec** (SAST, rapport SARIF) et **govulncheck**
  (CVE sur les dépendances effectivement atteignables par le code, pas
  seulement présentes dans `go.sum`) — tous bloquants dans le pipeline
  (`exit-code 1`).
- **Seuil de vigilance** : 0 CRITICAL/HIGH toléré en CI ; toute exception
  documentée doit avoir une date de remédiation.
- **Lien dette technique** : une vulnérabilité connue non corrigée est de la
  dette qui devient exponentiellement plus coûteuse à traiter avec le temps
  (montée de version bloquée par des changements cumulés, exploitation
  possible avant correction) — bloquer le merge tant qu'elle existe empêche
  cette dette de s'installer.

## 3. Maintenabilité (*Maintainability*) — dette technique SonarCloud

- **Mesure** : *code smells*, duplication de code et *maintainability
  rating* remontés par le scan SonarCloud unique du backend
  (`sonar.projectKey=JTR220_collector`), complété côté front par ESLint +
  Prettier (bloquants sur le lint, dette front documentée et suivie dans
  `eslint.config.js`).
- **Seuil de vigilance** : le *technical debt ratio* SonarCloud ne doit pas
  augmenter d'un commit à l'autre sur les modules touchés ; toute règle
  désactivée doit être justifiée en commentaire (déjà la convention adoptée,
  ex. les 3 règles Svelte désactivées documentées).
- **Lien dette technique** : c'est l'indicateur le plus direct — il *mesure*
  la dette technique elle-même (complexité cyclomatique excessive,
  duplication, code mort) plutôt qu'un symptôme, et permet d'agir avant
  qu'elle ne ralentisse les développements futurs.

## 4. Efficacité de performance (*Performance efficiency*) — comportement sous charge

- **Mesure** : temps de réponse moyen/p95 et taux d'erreur sous charge
  simulée (Siege ou JMeter) sur les endpoints critiques (`/article`,
  `/article/:id/buy`), démontré en soutenance.
- **Seuil de vigilance** : dégradation du p95 ou apparition d'erreurs 5xx à
  charge constante par rapport à une mesure de référence précédente.
- **Lien dette technique** : une dégradation de performance non détectée à
  temps (ex. requête N+1, absence d'index, pool de connexions DB non borné)
  s'aggrave avec la croissance du trafic et devient plus coûteuse à corriger
  une fois le code alentour construit sur l'hypothèse (fausse) que la
  requête est rapide. Le pool de connexions PostgreSQL est déjà borné
  (25 open / 5 idle / 30 min lifetime, tous services) précisément pour éviter
  ce type de dette de scalabilité.

## Synthèse

| Indicateur | Exigence ISO 25010 couverte | Outil | Bloquant en CI |
|---|---|---|---|
| Couverture de tests | Fiabilité | `go test -race` + coverage | Oui (échec test) |
| Vulnérabilités CRITICAL/HIGH | Sécurité | Trivy, gosec, govulncheck | Oui |
| Dette technique / code smells | Maintenabilité | SonarCloud, ESLint | Oui (lint) / suivi (Sonar) |
| Latence & erreurs sous charge | Efficacité de performance | Siege/JMeter | Non (démo ponctuelle) |

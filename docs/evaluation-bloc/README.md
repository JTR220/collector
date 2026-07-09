# Dossier de validation — Bloc « Superviser et assurer le développement des applications logicielles »

Ce dossier regroupe les livrables écrits demandés par les consignes CESI
(INFMAALSIA2V) qui n'avaient pas encore de support dédié dans le repo. Il
complète — sans le dupliquer — ce qui existe déjà :

- [`docs/DEVSECOPS.md`](../DEVSECOPS.md) : mapping des 15 étapes du pipeline
  DevSecOps/GitOps, schéma du flux CI/CD, état du bootstrap cluster.
- [`infra/README.md`](../../infra/README.md) : architecture GitOps (Kustomize,
  Argo CD, overlays staging/prod).
- Code + tests : `apps/backend/*/controllers/*_test.go` (unitaires),
  `apps/backend/*/routes/routes_integration_test.go` (intégration/acceptation).

## Sommaire

| Document | Critère de la grille couvert |
|---|---|
| [01-protocole-experimentation.md](01-protocole-experimentation.md) | Protocole d'expérimentation bac à sable (2 pts) |
| [02-backlog-fonctionnalite-metier.md](02-backlog-fonctionnalite-metier.md) | Fonctionnalité métier conforme au besoin MOA (1 pt) |
| [03-indicateurs-qualite.md](03-indicateurs-qualite.md) | 4 indicateurs qualité ISO 25010 vs dette technique (1 pt) |
| [04-cartographie-competences-formation.md](04-cartographie-competences-formation.md) | Compétences + action de formation (1 pt) |
| [05-plan-remediation-securite.md](05-plan-remediation-securite.md) | Plan de remédiation priorisé (2 pts) |
| [06-preparation-soutenance-jury.md](06-preparation-soutenance-jury.md) | Support de révision oral : architecture, sécurité, tests, GitOps, questions jury |

## Ce qui reste hors de ce dossier (à faire en dehors du code)

- **Démo de montée en charge** (Siege/JMeter) — dépend d'un cluster staging
  joignable en HTTPS au moment de la soutenance, à rejouer avant la date.
- **Schémas d'architecture** — la matière existe (services, flux CI/CD décrit
  textuellement dans `docs/DEVSECOPS.md`), mais aucune évaluation ne porte sur
  la forme d'un support de présentation ; à mettre en image le jour J si utile
  à l'oral.

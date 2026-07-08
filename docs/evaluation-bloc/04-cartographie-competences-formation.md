# Cartographie des compétences et action de formation

## Équipe (d'après le contexte Collector)

| Profil | Compétences déjà en place |
|---|---|
| Lead Developer (rôle tenu dans cet exercice) | Architecture microservices, Go/Gin, SvelteKit, CI/CD GitHub Actions, GitOps Argo CD/Kustomize, sécurité applicative de base (JWT, bcrypt, rate-limiting) |
| Développeur·euse confirmé·e #1 (5 ans d'XP) | Développement backend Go, tests unitaires, intégration API REST |
| Développeur·euse confirmé·e #2 (5 ans d'XP) | Développement frontend SvelteKit/TypeScript, intégration UI |

## Compétences nécessaires au projet

Recensées à partir des exigences du contexte (transactions financières,
recommandations, détection de fraude, internationalisation) et de
l'architecture cible :

1. **Sécurité applicative avancée** : gestion fine des autorisations
   (au-delà du JWT + rôle actuel), détection de fraude/anomalies, conformité
   paiement (protection contre l'échange de coordonnées personnelles entre
   utilisateurs, cf. `05-plan-remediation-securite.md`).
2. **Observabilité & exploitation Kubernetes** : instrumentation
   Prometheus des services Go (métriques applicatives, pas seulement
   infrastructure), lecture de dashboards Grafana, troubleshooting sur un
   cluster managé par GitOps.
3. **Data / recommandation** : l'exigence de recommandations basées sur les
   centres d'intérêt (V1) puis sur le parcours de navigation (V2) demande une
   compétence data science/ML absente de l'équipe actuelle.
4. **Accessibilité & internationalisation front** : compétence UX/a11y
   (WCAG) et i18n SvelteKit non couverte aujourd'hui.

## Écarts identifiés

- L'équipe maîtrise le développement applicatif et l'infrastructure GitOps
  (acquis via la mise en place déjà réalisée du pipeline DevSecOps), mais n'a
  pas de compétence dédiée en **sécurité offensive/détection de fraude** ni
  en **data/ML**, deux axes explicitement demandés par la direction
  (« outil d'automatisation de détection de fraude », « recommandations »).
- Pas de compétence a11y/i18n formalisée alors que ces exigences figurent
  explicitement dans le contexte.

## Action de formation proposée

**Formation : « Sécuriser une architecture microservices et détecter la
fraude applicative »** (format court, 2-3 jours, pour le Lead Developer et un
développeur confirmé), couvrant :

- Durcissement des API (OWASP API Security Top 10 appliqué à un contexte
  paiement/marketplace),
- Bases de détection d'anomalies applicables sans expertise ML lourde
  (règles métier, seuils statistiques simples — cohérent avec le
  `price-tracker-service` déjà en place qui applique déjà ce principe pour
  les pics de prix),
- Retour d'expérience sur la mise en Enforce de politiques Kyverno (actuellement
  en Audit) et la gestion de secrets en production.

Ce choix est réaliste au regard de la taille de l'équipe (pas de recrutement
d'un profil « mouton à 5 pattes » data-scientist-sécurité-DevOps) : il monte
en compétence l'équipe existante sur un axe directement actionnable à court
terme (sécurité), tout en laissant le sujet data/recommandation comme
évolution V2 nécessitant, elle, un recrutement dédié le moment venu.

# Schémas visuels — soutenance

Rendus PNG/SVG des diagrammes Mermaid (sources déjà présentes dans
[`../README.md`](../README.md) pour 01/03/04, complétées ici par 02 et 05).
Utiliser les `.png` (fond blanc) pour les coller dans un support de
présentation ; les `.svg` pour un affichage net à l'écran.

| Fichier | Contenu | Critère de la grille couvert |
|---|---|---|
| [01-architecture-systeme](01-architecture-systeme.png) | Vue système : front, 4 services Go, 3 bases Postgres, RabbitMQ | Architecture technique (4 pts) |
| [02-pipeline-cicd-gitops](02-pipeline-cicd-gitops.png) | Flux push → CI → image signée → Argo CD → promotion prod | Processus de livraison continue schématisé (3 pts) |
| [03-flux-achat](03-flux-achat.png) | Séquence d'achat (fonctionnalité métier du POC) | Fonctionnalité métier conforme au besoin MOA (1 pt) |
| [04-flux-prix-fraude](04-flux-prix-fraude.png) | Séquence changement de prix → détection fraude → notification temps réel | Illustration architecture événementielle |
| [05-cycle-devsecops](05-cycle-devsecops.png) | Cycle de vie DevSecOps en 6 étapes avec mesure de sécurité à chaque étape | Cycle de vie + sécurité du développement (§1.1.2 des consignes) |

## Régénérer après une modification du `.mmd`

```bash
npx -y @mermaid-js/mermaid-cli -i 0X-nom.mmd -o 0X-nom.svg -b transparent
npx -y @mermaid-js/mermaid-cli -i 0X-nom.mmd -o 0X-nom.png -b white -s 3
```

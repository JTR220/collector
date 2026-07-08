# Protocole d'expérimentation en bac à sable — HTTPS via cert-manager sur Kubernetes

## Technologie testée

**cert-manager** (opérateur Kubernetes de provisionnement automatique de
certificats TLS) couplé à un **Ingress NGINX** et à l'autorité **Let's
Encrypt**, en vue de sécuriser en HTTPS l'accès public à collector.shop
(exigence de sécurité de premier plan du contexte, l'application traitant des
transactions financières).

C'est un exemple explicitement cité comme valide par les consignes
(« déploiement d'un cert-manager pour le provisionnement de certificats
sécurisant un Ingress ») : il s'agit d'une plateforme support (l'opérateur
cert-manager + ses CRD `ClusterIssuer`/`Certificate`) et non d'un simple test
de langage ou de communication front↔back.

## Environnement de test

- Cluster **k3s** (Kubernetes léger single-node) sur un VPS OVH dédié
  (`vps-a6745acc.vps.ovh.net`, Ubuntu 26.04, 4 vCore/8 Go RAM) — voir
  `docs/DEVSECOPS.md` § « État du bootstrap cluster ».
- Argo CD (App-of-Apps, `infra/argocd/bootstrap/root-app.yaml`) pilote le
  déploiement déclaratif de `cert-manager` en tant qu'application gérée
  (`infra/argocd/apps/cert-manager.yaml`, `cert-manager-config.yaml`).
- DNS : sous-domaines `chaker.pro` pointés vers l'IP publique du VPS.
- Manifests testés : `infra/cert-manager/cluster-issuer.yaml` (ClusterIssuer
  Let's Encrypt), `infra/k8s/overlays/staging/ingress.yaml` +
  `infra/k8s/overlays/staging/certificate.yaml`, puis leur pendant prod
  (`infra/k8s/overlays/prod/ingress.yaml`).

## Étapes clés pour reproduire l'expérimentation

1. Bootstrap du cluster k3s + Argo CD (`infra/README.md`).
2. Déploiement de l'application Argo CD `cert-manager` (Helm chart officiel
   Jetstack) via l'App-of-Apps.
3. Application du `ClusterIssuer` référencant le serveur ACME Let's Encrypt
   (challenge HTTP-01, résolu par l'Ingress NGINX déjà en place).
4. Déclaration d'une ressource `Certificate` par overlay (staging/prod),
   référencée par l'`Ingress` du front (`ingressClassName: nginx`,
   `tls: - secretName: ... hosts: [...]`).
5. Vérification de l'émission : `kubectl get certificate -n <ns>` jusqu'à
   `READY=True`, puis test `curl -v https://<domaine>` (chaîne de
   certification valide, redirection HTTP→HTTPS active côté Ingress).

## Difficultés rencontrées

- **Résolution du challenge HTTP-01** : nécessite que le port 80 du VPS soit
  accessible publiquement et que l'Ingress route bien `/.well-known/acme-challenge/`
  vers le pod temporaire de cert-manager avant même que le certificat final
  n'existe — un ordre de déploiement incorrect (Ingress TLS actif avant
  l'émission du certificat) bloque la validation.
- **Propagation DNS** : le challenge échoue tant que le sous-domaine ne
  résout pas encore vers l'IP du VPS ; nécessite d'attendre la propagation
  avant de relancer l'émission (les CertificateRequest ont un TTL de retry
  limité chez Let's Encrypt, rate-limit à surveiller en cas de multiples
  essais infructueux).
- **Double environnement (staging/prod)** : deux `Certificate` distincts avec
  des noms de secret différents pour éviter qu'Argo CD ne les fasse
  s'écraser lors des synchronisations croisées d'overlays.

## Résultats obtenus

- Certificats émis et renouvelés automatiquement (renouvellement gardé par
  cert-manager ~30 jours avant expiration, sans intervention manuelle) pour
  les domaines `chaker.pro` en staging et prod.
- HTTPS opérationnel de bout en bout (Ingress → TLS termination → services
  internes en HTTP simple, cohérent avec un cluster mono-tenant de confiance).
- Aucune régression sur le flux HTTP existant (redirection 301 vers HTTPS
  conservée pour compatibilité des liens existants).

## Décision

**Adoption validée** : cert-manager + Let's Encrypt est retenu comme solution
de gestion des certificats pour collector.shop. Alternative écartée avant
expérimentation : certificats manuels (rotation humaine incompatible avec un
flux GitOps automatisé — la moindre erreur d'oubli de renouvellement casse la
disponibilité, contraire à l'exigence de fiabilité de la V1).

#!/usr/bin/env bash
# Test de montee en charge (indicateur #4, docs/evaluation-bloc/03-indicateurs-qualite.md)
# sur les endpoints publics en lecture de catalog-service (GET /health, /article,
# /category, /article/:id) : aucune ecriture, rejouable sans effet de bord ni
# risque d'epuiser le stock d'articles de demo.
#
# Usage :
#   ./run-siege.sh <base-url> [concurrence] [duree]
#
# Exemples :
#   ./run-siege.sh http://localhost:8081                 # local docker-compose
#   ./run-siege.sh https://staging.chaker.pro:8443/api    # staging (path /api)
#   ./run-siege.sh https://api.chaker.pro                 # prod (sous-domaine dedie)
set -euo pipefail

BASE_URL="${1:?Usage: run-siege.sh <base-url> [concurrence=25] [duree=1M]}"
CONCURRENCY="${2:-25}"
DURATION="${3:-1M}"

if ! command -v siege >/dev/null 2>&1; then
  echo "siege n'est pas installe. Sur le VPS (Ubuntu) : sudo apt install -y siege" >&2
  echo "En local (WSL/Debian) : sudo apt install -y siege" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
URLS_FILE="$(mktemp)"
sed "s#__BASE__#${BASE_URL}#g" "${SCRIPT_DIR}/urls.txt.tpl" > "${URLS_FILE}"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
REPORT="${SCRIPT_DIR}/results-${TIMESTAMP}.log"

echo "Cible        : ${BASE_URL}"
echo "Concurrence  : ${CONCURRENCY} utilisateurs simules"
echo "Duree        : ${DURATION}"
echo "URLs testees :"
cat "${URLS_FILE}"
echo "---"

siege --concurrent="${CONCURRENCY}" --time="${DURATION}" \
      --file="${URLS_FILE}" --internet --delay=1 \
      --log="${REPORT}" 2>&1 | tee "${SCRIPT_DIR}/console-${TIMESTAMP}.log"

rm -f "${URLS_FILE}"

echo "---"
echo "Rapport Siege : ${REPORT}"
echo "Sortie console : ${SCRIPT_DIR}/console-${TIMESTAMP}.log"
echo "A retenir pour la soutenance : transactions/sec, temps de reponse moyen,"
echo "taux de disponibilite (%) et nombre d'echecs (failed transactions)."

# Étape 5 — build multi-stage : toolchain Go isolée, image finale distroless
# static (pas de shell, pas de gestionnaire de paquets, ~2 Mo de base) en non-root.
FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog-service . \
    && mkdir -p /out/uploads/articles

# static-debian12 suffit : binaire statique (CGO désactivé), CA certs et tzdata inclus
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/catalog-service /app/catalog-service
# le service écrit les photos d'articles dans ./uploads : le dossier doit
# appartenir à nonroot (uid 65532) — en k8s, le volume est monté par-dessus
COPY --from=builder --chown=65532:65532 /out/uploads /app/uploads

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/app/catalog-service"]

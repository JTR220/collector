# Étape 5 — build multi-stage : toolchain Go isolée, image finale distroless
# static (pas de shell, pas de gestionnaire de paquets, ~2 Mo de base) en non-root.
FROM golang:1.26 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notification-service ./cmd/main.go

# static-debian12 suffit : binaire statique (CGO désactivé), CA certs et tzdata inclus.
# Pas de HEALTHCHECK : distroless n'a pas de wget, les probes Kubernetes s'en chargent.
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/notification-service /app/notification-service

EXPOSE 8083

USER nonroot:nonroot

ENTRYPOINT ["/app/notification-service"]

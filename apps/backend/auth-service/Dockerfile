# Étape 5 — build multi-stage : toolchain Go isolée, image finale distroless
# static (pas de shell, pas de gestionnaire de paquets, ~2 Mo de base) en non-root.
FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/auth-service .

# static-debian12 suffit : binaire statique (CGO désactivé), CA certs et tzdata inclus
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/auth-service /app/auth-service

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/auth-service"]

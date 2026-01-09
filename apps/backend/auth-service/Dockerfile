
FROM golang:1.23-alpine AS builder


WORKDIR /app

COPY go.mod go.sum ./
# On télécharge les dépendances (mise en cache pour aller plus vite les prochaines fois)
RUN go mod download

# On copie tout le code source
COPY . .

# On compile l'application en un exécutable binaire nommé "main"
RUN go build -o main cmd/api/main.go

# -----------------------------------------------

# Étape 2 : Le Runner (L'image finale)
# On repart d'une image vide (Alpine) pour que ce soit très léger
FROM alpine:latest

WORKDIR /root/

# On copie uniquement l'exécutable créé à l'étape 1
COPY --from=builder /app/main .
# On copie aussi le fichier .env (attention, en prod on fera autrement, mais pour l'instant c'est ok)
COPY .env .

# On expose le port 8080
EXPOSE 8080

# La commande qui se lance au démarrage du conteneur
CMD ["./main"]
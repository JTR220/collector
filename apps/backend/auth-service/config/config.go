// Package config centralise la lecture des variables d'environnement du
// service, avec des valeurs par defaut saines pour le developpement local.
package config

import (
	"os"
	"strconv"
)

// EnvOr renvoie la variable d'environnement si elle est definie, sinon def.
func EnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// EnvInt renvoie la variable d'environnement convertie en entier si elle est
// definie et valide, sinon def.
func EnvInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

// EnvBool renvoie true si la variable vaut "true" (insensible a la casse non
// geree volontairement : la convention du projet est la minuscule).
func EnvBool(key string) bool {
	return os.Getenv(key) == "true"
}

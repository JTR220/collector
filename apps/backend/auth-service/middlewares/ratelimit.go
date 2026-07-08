package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit limite chaque IP a max requetes par fenetre glissante window.
// Implementation en memoire, suffisante pour un service mono-instance :
// protege les endpoints d'authentification (login, inscription) contre le
// brute force et l'enumeration de comptes.
func RateLimit(max int, window time.Duration) gin.HandlerFunc {
	var (
		mu   sync.Mutex
		hits = map[string][]time.Time{}
	)

	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()

		mu.Lock()
		// Purge globale occasionnelle pour borner la memoire si beaucoup
		// d'IP distinctes ont ete vues.
		if len(hits) > 10_000 {
			for k, stamps := range hits {
				if len(stamps) == 0 || now.Sub(stamps[len(stamps)-1]) >= window {
					delete(hits, k)
				}
			}
		}

		recent := hits[ip][:0]
		for _, ts := range hits[ip] {
			if now.Sub(ts) < window {
				recent = append(recent, ts)
			}
		}

		if len(recent) >= max {
			hits[ip] = recent
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Trop de tentatives, reessayez dans quelques instants"})
			return
		}

		hits[ip] = append(recent, now)
		mu.Unlock()

		c.Next()
	}
}

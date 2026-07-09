package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// purgeStaleHits borne la memoire occupee par la fenetre glissante : si
// beaucoup d'IP distinctes ont ete vues, on efface celles dont la derniere
// requete est deja hors fenetre.
func purgeStaleHits(hits map[string][]time.Time, now time.Time, window time.Duration) {
	if len(hits) <= 10_000 {
		return
	}
	for ip, stamps := range hits {
		if len(stamps) == 0 || now.Sub(stamps[len(stamps)-1]) >= window {
			delete(hits, ip)
		}
	}
}

// recentHits ne garde, parmi les horodatages d'une IP, que ceux encore dans
// la fenetre glissante.
func recentHits(stamps []time.Time, now time.Time, window time.Duration) []time.Time {
	recent := stamps[:0]
	for _, ts := range stamps {
		if now.Sub(ts) < window {
			recent = append(recent, ts)
		}
	}
	return recent
}

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
		purgeStaleHits(hits, now, window)
		recent := recentHits(hits[ip], now, window)

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

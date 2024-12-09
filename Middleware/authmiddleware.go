package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// SecretKey pour signer et vérifier les JWT
var SecretKey = []byte("test_cle_secrete")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			//log.Printf("Erreur lors de la récupération du cookie : %v", err)
			w.Header().Set("HX-Redirect", "/auth") // Redirige vers la page d'accueil
			w.WriteHeader(http.StatusOK)
			return
		}

		// La valeur brute du cookie
		rawValue := cookie.Value
		//	log.Printf("Valeur brute du cookie : %s", rawValue)

		// Reformater la valeur du cookie pour qu'elle soit un JSON valide
		// Ajouter des guillemets autour de la clé "token" si nécessaire
		rawValue = strings.Replace(rawValue, "token:", "", 1)
		rawValue = strings.Replace(rawValue, "{", "", 1)
		rawValue = strings.Replace(rawValue, "}", "", 1)
		tokenStr := rawValue
		///log.Printf("Token extrait : %s", tokenStr)

		// Décoder et valider le JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Vérifiez le type d'algorithme
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("méthode de signature inattendue : %v", token.Header["alg"])
			}
			return []byte("test_cle_secrete"), nil
		})

		if err != nil {
			//	log.Printf("Erreur lors de la validation du token : %v", err)
			w.Header().Set("HX-Redirect", "/auth") // Redirige vers la page d'accueil
			w.WriteHeader(http.StatusOK)
			return
		}

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			//log.Printf("Token valide : %v", claims)
			// Ajouter les informations du token dans le contexte si nécessaire
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("HX-Redirect", "/auth") // Redirige vers la page d'accueil
			w.WriteHeader(http.StatusOK)
		}
	})
}

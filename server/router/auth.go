package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chukfi/backend/database/schema"
	"github.com/chukfi/backend/src/httpresponder"
	"github.com/chukfi/backend/src/lib/permissions"
	"github.com/go-chi/chi/v5"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterAuthRoutes(r chi.Router, database *gorm.DB) {
	r.Route("/auth", func(r chi.Router) {

		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromRequest(r, database)

			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			type simpleUser struct {
				ID          string   `json:"id"`
				Fullname    string   `json:"fullname"`
				Email       string   `json:"email"`
				Permissions []string `json:"permissions"`
			}

			perms := permissions.PermissionsToStrings(permissions.Permission(user.Permissions))

			httpresponder.SendNormalResponse(w, r, map[string]interface{}{
				"user": simpleUser{
					ID:          user.ID.String(),
					Fullname:    user.Fullname,
					Email:       user.Email,
					Permissions: perms,
				},
				"success": true,
			})
		})

		r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
			// check if already logged in
			authToken, ok := r.Context().Value("authToken").(string)
			if ok && authToken != "" {
				// check if token is valid
				result, err := gorm.G[schema.UserToken](database).Where("token = ? AND expires_at > ?", authToken, time.Now().Unix()).First(r.Context())
				if err == nil && result.ID != uuid.Nil {
					httpresponder.SendErrorResponse(w, r, "Already logged in", http.StatusBadRequest)
					return
				}
			}

			var body loginRequest
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Invalid request body: "+err.Error(), http.StatusBadRequest)
				return
			}

			if body.Email == "" || body.Password == "" {
				httpresponder.SendErrorResponse(w, r, "Email and password are required", http.StatusBadRequest)
				return
			}

			var user schema.User
			result := database.Where("email = ?", body.Email).First(&user)
			if result.Error != nil {
				httpresponder.SendErrorResponse(w, r, "Invalid email or password", http.StatusUnauthorized)
				return
			}

			// bcrypt compare
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Invalid email or password", http.StatusUnauthorized)
				return
			}
			// create auth token
			token := uuid.NewV4()

			userToken := schema.UserToken{
				UserID:    user.ID,
				Token:     token.String(),
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			}

			err = gorm.G[schema.UserToken](database).Create(r.Context(), &userToken)
			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Failed to save auth token: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// set cookie
			http.SetCookie(w, &http.Cookie{
				Name:    "chukfi_auth_token",
				Value:   token.String(),
				Expires: time.Unix(userToken.ExpiresAt, 0),
				Path:    "/",
			})

			type simpleUser struct {
				ID          string   `json:"id"`
				Fullname    string   `json:"fullname"`
				Email       string   `json:"email"`
				Permissions []string `json:"permissions"`
			}

			perms := permissions.PermissionsToStrings(permissions.Permission(user.Permissions))

			httpresponder.SendNormalResponse(w, r, map[string]interface{}{
				"authToken": token.String(),
				"expiresAt": userToken.ExpiresAt,
				"user": simpleUser{
					ID:          user.ID.String(),
					Fullname:    user.Fullname,
					Email:       user.Email,
					Permissions: perms,
				},
				"success": true,
			})
		})
	})
}

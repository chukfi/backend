package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chukfi/backend/database/schema"
	"github.com/chukfi/backend/src/chumiddleware"
	"github.com/chukfi/backend/src/httpresponder"
	usercache "github.com/chukfi/backend/src/lib/cache/user"
	"github.com/chukfi/backend/src/lib/permissions"
	"github.com/chukfi/backend/src/lib/schemaregistry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

/*
GetUserIDFromAuthToken retrieves the userID associated with the given auth token from the database.
*/
func GetUserIDFromAuthToken(database *gorm.DB, authToken string) (string, error) {
	var userToken schema.UserToken
	result := database.Where("token = ? AND expires_at > ?", authToken, time.Now().Unix()).First(&userToken)
	if result.Error != nil {
		return "", result.Error
	}
	return userToken.UserID.String(), nil
}

/*
GetUserIDFromRequest retrieves the userID from the request context. If not found, it returns an empty string.
*/
func GetUserIDFromRequest(request *http.Request) string {
	userID, ok := request.Context().Value("userID").(string)
	if !ok || userID == "" {
		return ""
	}
	return userID
}

/*
GetUserFromRequest retrieves the user associated with the request. It first checks the request context for a userID.
If not found, it looks for an auth token in the context, validates it against the database,
and retrieves the corresponding user.
*/
func GetUserFromRequest(request *http.Request, database *gorm.DB) (*schema.User, error) {
	userID, ok := request.Context().Value("userID").(string)
	if !ok || userID == "" {
		// try to get from auth token
		authToken, ok := request.Context().Value("authToken").(string)
		if !ok || authToken == "" {
			return nil, fmt.Errorf("no user ID or auth token in request")
		}

		// check cache
		var result, err = gorm.G[schema.UserToken](database).Where("token = ? AND expires_at > ?", authToken, time.Now().Unix()).First(request.Context())
		if err != nil || result.ID == uuid.Nil {
			return nil, fmt.Errorf("invalid auth token")
		}
		userID = result.UserID.String()
	}

	cacheduser, found := usercache.UserCacheInstance.Get(userID)

	if found {
		return &cacheduser, nil
	}

	var user schema.User
	result := database.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	usercache.UserCacheInstance.Set(userID, user)
	return &user, nil
}

/*
RequestRequiresPermission checks if the user associated with the request has the required permissions.
*/
func RequestRequiresPermission(request *http.Request, database *gorm.DB, requiredPermissions permissions.Permission) bool {
	user, err := GetUserFromRequest(request, database)
	if err != nil {
		return false
	}

	return permissions.HasPermission(permissions.Permission(user.Permissions), requiredPermissions)

}

/*
RoutesRequiresPermission is a middleware that checks if the user has the required permissions to access the route.
If not, it returns a 403 Forbidden response.
*/
func RoutesRequiresPermission(database *gorm.DB, required permissions.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user, err := GetUserFromRequest(r, database)
			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
			if !permissions.HasPermission(permissions.Permission(user.Permissions), required) {
				httpresponder.SendErrorResponse(w, r, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

/*
AuthMiddlewareWithDatabase checks for auth token in context, validates it against the database,
and if valid, adds the userID to the request context.
*/
func AuthMiddlewareWithDatabase(database *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if database == nil {
				httpresponder.SendErrorResponse(w, r, "Database not initialized", http.StatusInternalServerError)
				return
			}

			authToken, ok := r.Context().Value("authToken").(string)

			if !ok || authToken == "" {
				httpresponder.SendErrorResponse(w, r, "Unauthorized: No auth token provided", http.StatusUnauthorized)
				return
			}

			result, err := gorm.G[schema.UserToken](database).Where("token = ? AND expires_at > ?", authToken, time.Now().Unix()).First(r.Context())

			if err != nil {
				httpresponder.SendErrorResponse(w, r, "Unauthorized: Invalid auth token", http.StatusUnauthorized)
				return
			}

			if result.ExpiresAt < time.Now().Unix() {
				httpresponder.SendErrorResponse(w, r, "Unauthorized: Auth token expired", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", result.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetupRouter(database *gorm.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(chumiddleware.CaseSensitiveMiddleware)
	r.Use(chumiddleware.SaveAuthTokenMiddleware)

	// admin routes with database so /admin/collection/${collectionName}/get

	r.Route("/admin", func(r chi.Router) {
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

		r.Route("/collection", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(AuthMiddlewareWithDatabase(database))

				r.Get("/all", func(w http.ResponseWriter, r *http.Request) {
					// gets all
					user, err := GetUserFromRequest(r, database)

					if err != nil {
						httpresponder.SendErrorResponse(w, r, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
						return
					}

					hasPermission := permissions.HasPermission(permissions.Permission(user.Permissions), permissions.ViewModels)
					if !hasPermission {
						httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to view models", http.StatusForbidden)
						return
					}

					allSchemas := schemaregistry.GetAllRegisteredSchemas()

					httpresponder.SendNormalResponse(w, r, map[string]interface{}{
						"schemas": allSchemas,
					})

				})
			})
			r.Route("/{collectionName}", func(r chi.Router) {

				// auth mandated routes

				r.Group(func(r chi.Router) {
					r.Use(AuthMiddlewareWithDatabase(database))

					r.Get("/metadata", func(w http.ResponseWriter, r *http.Request) {
						hasPermission := RequestRequiresPermission(r, database, permissions.ViewModels)
						if !hasPermission {
							httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection metadata", http.StatusForbidden)
							return
						}

						collectionName := chi.URLParam(r, "collectionName")

						resolvedName, exists := schemaregistry.ResolveTableName(collectionName)
						if !exists {
							httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
							return
						}
						collectionName = resolvedName

						metadata, _ := schemaregistry.GetMetadata(collectionName)

						httpresponder.SendNormalResponse(w, r, metadata)
					})

					r.Post("/create", func(w http.ResponseWriter, r *http.Request) {
						// this route creates a new entry in the specified collection
						collectionName := chi.URLParam(r, "collectionName")

						// resolve collection name (allows singular or plural)
						resolvedName, exists := schemaregistry.ResolveTableName(collectionName)
						if !exists {
							httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
							return
						}
						collectionName = resolvedName

						if schemaregistry.IsAdminOnly(collectionName) {
							hasPermission := RequestRequiresPermission(r, database, permissions.ManageModels)
							if !hasPermission {
								httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection metadata", http.StatusForbidden)
								return
							}
						}

						// parse body into map
						var data map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&data)
						if err != nil {
							httpresponder.SendErrorResponse(w, r, "Invalid request body: "+err.Error(), http.StatusBadRequest)
							return
						}

						missing, unknown := schemaregistry.ValidateBody(collectionName, data)

						if len(missing) > 0 {
							httpresponder.SendErrorResponse(w, r, "Missing required fields: "+strings.Join(missing, ", "), http.StatusBadRequest)
							return
						}

						if len(unknown) > 0 {
							httpresponder.SendErrorResponse(w, r, "Unknown fields: "+strings.Join(unknown, ", "), http.StatusBadRequest)
							return
						}

						data["ID"] = uuid.NewV4()
						data["created_at"] = time.Now()
						data["updated_at"] = time.Now()

						result := gorm.G[map[string]interface{}](database).Table(collectionName).Create(r.Context(), &data)

						if result != nil {
							httpresponder.SendErrorResponse(w, r, "Error creating entry: "+result.Error(), http.StatusInternalServerError)
							return
						}

						httpresponder.SendNormalResponse(w, r, data)

					})
				})

				// non auth mandated routes
				r.Get("/get", func(w http.ResponseWriter, r *http.Request) {
					collectionName := chi.URLParam(r, "collectionName")

					// resolve collection name (allows singular or plural)
					resolvedName, exists := schemaregistry.ResolveTableName(collectionName)
					if !exists {
						httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
						return
					}
					collectionName = resolvedName

					if schemaregistry.IsAdminOnly(collectionName) {
						// check auth
						hasPermission := RequestRequiresPermission(r, database, permissions.ManageModels)
						if !hasPermission {
							httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection metadata", http.StatusForbidden)
							return
						}
					}

					var results []map[string]interface{}

					result, err := gorm.G[map[string]interface{}](database).Table(collectionName).Find(r.Context())
					if err != nil {
						if err == gorm.ErrRecordNotFound {
							httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
							return
						}
						httpresponder.SendErrorResponse(w, r, "Error fetching collection: "+err.Error(), http.StatusInternalServerError)
						return
					}
					results = result

					httpresponder.SendNormalResponse(w, r, results)
				})
			})

		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
	})

	return r
}

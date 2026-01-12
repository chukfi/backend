package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chukfi/backend/database/schema"
	"github.com/chukfi/backend/src/chumiddleware"
	"github.com/chukfi/backend/src/httpresponder"
	usercache "github.com/chukfi/backend/src/lib/cache/user"
	"github.com/chukfi/backend/src/lib/permissions"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

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
	userPerms := permissions.Permission(user.Permissions)

	result := permissions.HasPermission(userPerms, requiredPermissions)
	return result
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

			ctx := context.WithValue(r.Context(), "userID", result.UserID.String())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetupRouter(database *gorm.DB, frontendDirectory ...string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(chumiddleware.CaseSensitiveMiddleware)
	r.Use(chumiddleware.SaveAuthTokenMiddleware)

	// if frontendDirectory is set, serve static files from there
	if len(frontendDirectory) > 0 && frontendDirectory[0] != "" {
		fileServer := http.FileServer(http.Dir(frontendDirectory[0]))
		r.Handle("/*", http.StripPrefix("/", fileServer))
	} else {
		yellow := "\033[33m"
		reset := "\033[0m"
		fmt.Println(string(yellow), "Warning: Frontend directory not set. Static files will not be served.", string(reset))
	}

	// admin routes with database so /admin/collection/${collectionName}/get

	r.Route("/admin", func(r chi.Router) {
		RegisterAuthRoutes(r, database)
		RegisterCollectionRoutes(r, database)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
	})

	return r
}

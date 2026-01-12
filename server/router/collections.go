package router

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/chukfi/backend/src/httpresponder"
	"github.com/chukfi/backend/src/lib/permissions"
	"github.com/chukfi/backend/src/lib/schemaregistry"
	"github.com/go-chi/chi/v5"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

func RegisterCollectionRoutes(r chi.Router, database *gorm.DB) {
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

				// allows updating entries in collections
				r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
					// this route updates an entry in the specified collection
					collectionName := chi.URLParam(r, "collectionName")

					// resolve collection name (allows singular or plural)
					resolvedName, exists := schemaregistry.ResolveTableName(collectionName)
					if !exists {
						httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
						return
					}
					collectionName = resolvedName

					hasPermission := RequestRequiresPermission(r, database, permissions.ManageModels)
					if !hasPermission {
						httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection metadata", http.StatusForbidden)
						return
					}

					// parse body into map
					var data map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&data)
					if err != nil {
						httpresponder.SendErrorResponse(w, r, "Invalid request body: "+err.Error(), http.StatusBadRequest)
						return
					}

					idValue, exists := data["ID"] 
					if !exists {
						httpresponder.SendErrorResponse(w, r, "Missing ID field in request body", http.StatusBadRequest)
						return
					}

					idStr, ok := idValue.(string)
					if !ok {
						httpresponder.SendErrorResponse(w, r, "Invalid ID field in request body", http.StatusBadRequest)
						return
					}

					id, err := uuid.FromString(idStr)
					if err != nil {
						httpresponder.SendErrorResponse(w, r, "Invalid ID format: "+err.Error(), http.StatusBadRequest)
						return
					}

					// check if everything else is valid to IsBodyMostlyValid
					isValid, err := schemaregistry.IsBodyMostlyValid(collectionName, data)
					if !isValid {
						httpresponder.SendErrorResponse(w, r, "Invalid request body: "+err.Error(), http.StatusBadRequest)
						return
					}

					// set updated_at
					data["updated_at"] = time.Now()

					res, err := gorm.G[map[string]interface{}](database).Table(collectionName).Where("id = ?", id).Updates(r.Context(), data)

					if err != nil {
						httpresponder.SendErrorResponse(w, r, "Error updating entry: "+err.Error(), http.StatusInternalServerError)
						return
					}

					if res == 0 {
						httpresponder.SendErrorResponse(w, r, "No entry found with the given ID", http.StatusBadRequest)
						return
					}

					httpresponder.SendNormalResponse(w, r, map[string]interface{}{
						"success": true,
					})
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

					hasPermission := RequestRequiresPermission(r, database, permissions.ManageModels)
					if !hasPermission {
						httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection metadata", http.StatusForbidden)
						return
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

			r.Post("/get", func(w http.ResponseWriter, r *http.Request) {
				collectionName := chi.URLParam(r, "collectionName")

				resolvedName, exists := schemaregistry.ResolveTableName(collectionName)
				if !exists {
					httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
					return
				}
				collectionName = resolvedName

				if schemaregistry.IsAdminOnly(collectionName) {
					authToken, ok := r.Context().Value("authToken").(string)
					if !ok || authToken == "" {
						httpresponder.SendErrorResponse(w, r, "Unauthorized: Authentication required for this collection", http.StatusUnauthorized)
						return
					}
					hasPermission := RequestRequiresPermission(r, database, permissions.ViewModels)
					if !hasPermission {
						httpresponder.SendErrorResponse(w, r, "Forbidden: You do not have permission to access this collection", http.StatusForbidden)
						return
					}
				}

				var body struct {
					Take   *int   `json:"take"`
					Page   *int   `json:"page"`
					Select string `json:"select"`
					Where  string `json:"where"`
				}
				json.NewDecoder(r.Body).Decode(&body)

				take := 30
				if body.Take != nil {
					take = *body.Take
					take = int(min(int64(take), 30)) // max 30
					if take <= 0 {
						take = 30
					}
				}

				page := 1
				if body.Page != nil && *body.Page > 0 {
					page = *body.Page
				}
				offset := (page - 1) * take

				query := database.Table(collectionName)

				if body.Select != "" {
					fields := strings.Split(body.Select, ",")
					for i := range fields {
						fields[i] = strings.TrimSpace(fields[i])
					}
					query = query.Select(fields)
				}

				if body.Where != "" {
					conditions := strings.Split(body.Where, ",")
					for _, condition := range conditions {
						parts := strings.SplitN(condition, ":", 2)
						if len(parts) == 2 {
							field := strings.TrimSpace(parts[0])
							value := strings.TrimSpace(parts[1])
							query = query.Where(field+" = ?", value)
						}
					}
				}

				query = query.Limit(take).Offset(offset)

				var results []map[string]interface{}
				err := query.Find(&results).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						httpresponder.SendErrorResponse(w, r, "Invalid collection name: "+collectionName, http.StatusBadRequest)
						return
					}
					httpresponder.SendErrorResponse(w, r, "Error fetching collection: "+err.Error(), http.StatusInternalServerError)
					return
				}

				httpresponder.SendNormalResponse(w, r, results)
			})
		})

	})
}

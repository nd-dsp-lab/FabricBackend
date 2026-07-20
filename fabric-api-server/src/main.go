package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-huma-api-server/src/auth"
	"go-huma-api-server/src/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"

	_ "github.com/danielgtaylor/huma/v2/formats/cbor"
)

type Options struct {
	Port      int  `help:"Port to listen on" short:"p" default:"8001"`
	LocalTest bool `help:"Run in local test mode without Fabric connection" short:"l" default:"false"`
}

type RecordInput struct {
	Body struct {
		RecordID   string `json:"recordID" doc:"Record ID, must be unique. If not provided, it will be generated automatically."`
		DroneID    string `json:"droneID" doc:"Drone ID"`
		PilotID    string `json:"pilotID" doc:"Pilot ID"`
		ZoneID     string `json:"zoneID" doc:"Zone ID"`
		RecordType string `json:"recordType" doc:"Record Type"`
		Reserved   string `json:"reserved" doc:"Reserved, a serialized JSON string"`
	}
}

type CertificateInput struct {
	Body struct {
		RecordID string `json:"recordID" doc:"Record ID"`
		DroneID  string `json:"droneID" doc:"Drone ID"`
		PilotID  string `json:"pilotID" doc:"Pilot ID"`
		ZoneID   string `json:"zoneID" doc:"Zone ID"`
		Reserved string `json:"reserved" doc:"Reserved, a serialized JSON string"`
	}
}

type ProfileInput struct {
	Body struct {
		RecordID string `json:"recordID" doc:"Record ID"`
		DroneID  string `json:"droneID" doc:"Drone ID"`
		PilotID  string `json:"pilotID" doc:"Pilot ID"`
		ZoneID   string `json:"zoneID" doc:"Zone ID"`
		Reserved string `json:"reserved" doc:"Reserved, a serialized JSON string"`
	}
}

type ReceviedRecordResponse struct {
	Body struct {
		Message string       `json:"message" doc:"Message"`
		Record  utils.Record `json:"record" doc:"Record"`
	}
}

type RecordIDInput struct {
	Body struct {
		RecordID string `json:"recordID" doc:"Record ID"`
	}
}

type RecordListResponse struct {
	Body struct {
		Message string         `json:"message" doc:"Message"`
		Records []utils.Record `json:"records"`
	} `json:"body"`
}

type CertificateSelector struct {
	Body struct {
		Selector map[string]interface{} `json:"selector" doc:"Selector to query certificates" default:"{\"recordType\": \"certificate\", \"pilotID\": \"Pilot00\"}"`
	}
}

type RecordSelector struct {
	Body struct {
		Selector map[string]interface{} `json:"selector" doc:"Selector to query records" default:"{\"droneID\": \"drone1\"}"`
	}
}

type ProfileSelector struct {
	Body struct {
		Selector map[string]interface{} `json:"selector" doc:"Selector to query profiles" default:"{\"recordType\": \"profile\", \"pilotID\": \"Pilot00\"}"`
	}
}

type ProfileUpdateInput struct {
	Body struct {
		RecordID   string                 `json:"recordID" doc:"Record ID"`
		UpdateInfo map[string]interface{} `json:"updateInfo" doc:"Info to update the profile" default:"{\"droneID\": \"drone1\", \"pilotID\": \"Pilot00\", \"zoneID\": \"Zone00\", \"reserved\": \"\"}"`
	}
}

type HistoryQueryResult struct {
	TxId      string        `json:"txId"`
	Timestamp time.Time     `json:"timestamp"`
	IsDelete  bool          `json:"isDelete"`
	Record    *utils.Record `json:"record"`
}

type HistoryQueryResultList struct {
	Body struct {
		Message string               `json:"message" doc:"Message"`
		History []HistoryQueryResult `json:"history" doc:"History of the record"`
	}
}

// ==================== Hash API types ====================

type HashStoreInput struct {
	Body struct {
		Identifier string `json:"identifier" doc:"Unique identifier for the hash entry"`
		Hash       string `json:"hash" doc:"Hash value to store"`
	}
}

type HashStoreResponse struct {
	Body struct {
		Message string `json:"message" doc:"Status message"`
	}
}

type HashRetrieveInput struct {
	Body struct {
		Identifier string `json:"identifier" doc:"Identifier to look up"`
	}
}

type HashRetrieveResponse struct {
	Body struct {
		Identifier string `json:"identifier" doc:"The identifier"`
		Hash       string `json:"hash" doc:"The stored hash value"`
	}
}

type EmptyInput struct{}

func main() {
	// Create a CLI app which takes a port option.
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		// Initialize Gateway only if not in local test mode
		if !options.LocalTest {
			utils.InitGateway()
		}

		// Initialize token store
		tokenFilePath := os.Getenv("TOKEN_FILE_PATH")
		if tokenFilePath == "" {
			tokenFilePath = "./tokens.json"
		}
		tokenStore := auth.NewTokenStore(tokenFilePath)
		if err := tokenStore.LoadTokens(); err != nil {
			log.Printf("Warning: Failed to load tokens: %v", err)
		}
		// Seed a default token for local testing
		const devToken = "dev-test-token"
		tokenStore.AddToken(devToken, "dev-secret-key")
		log.Printf("Loaded %d tokens from %s", tokenStore.GetTokenCount(), tokenFilePath)

		// Create a new router & API
		router := chi.NewMux()
		config := huma.DefaultConfig("My API", "1.0.0")
		config.Servers = []*huma.Server{
			{URL: "http://localhost:8001"},
			{URL: "https://proxy.web3db.org/nasa-project"},
		}
		// Add Bearer token security scheme for OpenAPI documentation
		config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "token",
				Description:  "Enter your Bearer token. Default testing token: dev-test-token",
			},
		}

		// Create Huma API from router
		api := humachi.New(router, config)

		// Create authenticated router group
		// We need to use Chi's Mount with a sub-router that has the auth middleware
		authRouter := chi.NewRouter()
		authRouter.Use(auth.BearerAuthMiddleware(tokenStore))

		// Create Huma API for authenticated routes
		authAPI := humachi.New(authRouter, config)

		// Mount the authenticated router
		router.Mount("/", authRouter)

		// Register POST /create-record (authenticated)
		huma.Register(authAPI, huma.Operation{
			OperationID:   "post-create-record",
			Method:        http.MethodPost,
			Path:          "/create-record",
			Summary:       "Create a record",
			Tags:          []string{"Records"},
			Description:   "Create a record.",
			DefaultStatus: http.StatusCreated,
			Security:      []map[string][]string{{"bearerAuth": {}}},
		}, func(ctx context.Context, input *RecordInput) (*ReceviedRecordResponse, error) {
			// Print secret key for authenticated request
			if authCtx := auth.GetAuthContext(ctx); authCtx != nil {
				fmt.Printf("[Auth] Secret key for token: %s\n", authCtx.SecretKey)
			}

			record := utils.Record{
				RecordID:   input.Body.RecordID,
				DroneID:    input.Body.DroneID,
				PilotID:    input.Body.PilotID,
				ZoneID:     input.Body.ZoneID,
				RecordType: input.Body.RecordType,
				Reserved:   input.Body.Reserved,
			}
			err := utils.CreateRecord(record.RecordID, record.DroneID, record.PilotID, record.ZoneID, record.RecordType, record.Reserved, options.LocalTest)
			if err != nil {
				return nil, err
			}
			resp := &ReceviedRecordResponse{}
			resp.Body.Message = "Record created successfully."
			resp.Body.Record = record
			return resp, nil
		})

		// Register POST /certificates/create (authenticated)
		huma.Register(authAPI, huma.Operation{
			OperationID:   "post-certificate-create",
			Method:        http.MethodPost,
			Path:          "/certificates/create",
			Summary:       "Post a certificate",
			Tags:          []string{"Certificates"},
			Description:   "Post a certificate to the blockchain.",
			DefaultStatus: http.StatusCreated,
			Security:      []map[string][]string{{"bearerAuth": {}}},
		}, func(ctx context.Context, input *CertificateInput) (*ReceviedRecordResponse, error) {
			if input.Body.RecordID == "" {
				input.Body.RecordID = utils.GetRandomString(10)
			}
			err := utils.CreateRecord(input.Body.RecordID, input.Body.DroneID, input.Body.PilotID, input.Body.ZoneID, "certificate", input.Body.Reserved, options.LocalTest)
			if err != nil {
				return nil, err
			}

			// Generate the response
			resp := &ReceviedRecordResponse{}
			resp.Body.Message = "Certificate received and stored successfully."
			resp.Body.Record = utils.Record{
				RecordID:   input.Body.RecordID,
				DroneID:    input.Body.DroneID,
				PilotID:    input.Body.PilotID,
				ZoneID:     input.Body.ZoneID,
				RecordType: "certificate",
				Reserved:   input.Body.Reserved,
			}
			return resp, nil
		})

		// Register POST /profiles/create (authenticated)
		huma.Register(authAPI, huma.Operation{
			OperationID:   "post-profile-create",
			Method:        http.MethodPost,
			Path:          "/profiles/create",
			Summary:       "Create a profile",
			Tags:          []string{"Profiles"},
			Description:   "Create a profile.",
			DefaultStatus: http.StatusCreated,
			Security:      []map[string][]string{{"bearerAuth": {}}},
		}, func(ctx context.Context, input *ProfileInput) (*ReceviedRecordResponse, error) {
			if input.Body.RecordID == "" {
				input.Body.RecordID = utils.GetRandomString(10)
			}
			err := utils.CreateRecord(input.Body.RecordID, input.Body.DroneID, input.Body.PilotID, input.Body.ZoneID, "profile", input.Body.Reserved, options.LocalTest)
			if err != nil {
				err = fmt.Errorf("failed to create record: %v, check if the recordID already exists", err)
				return nil, err
			}

			resp := &ReceviedRecordResponse{}
			resp.Body.Message = "Profile received and stored successfully."
			resp.Body.Record = utils.Record{
				RecordID:   input.Body.RecordID,
				DroneID:    input.Body.DroneID,
				PilotID:    input.Body.PilotID,
				ZoneID:     input.Body.ZoneID,
				RecordType: "profile",
				Reserved:   input.Body.Reserved,
			}
			return resp, nil
		})

		// Register POST /certificates/query
		huma.Register(api, huma.Operation{
			OperationID:   "post-certificate-query",
			Method:        http.MethodPost,
			Path:          "/certificates/query",
			Summary:       "Query certificates with selector",
			Tags:          []string{"Certificates"},
			Description:   "Query certificates with a standard mango selector.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *CertificateSelector) (*RecordListResponse, error) {
			// log.Printf("Received selector: %v", input.Body.Selector)

			// Convert selector map to JSON string
			selectorJSON, err := json.Marshal(map[string]interface{}{"selector": input.Body.Selector})
			if err != nil {
				return nil, err
			}

			recordsJSON, err := utils.GetRecordWithSelector(string(selectorJSON), options.LocalTest)
			if err != nil {
				return nil, err
			}

			// Parse JSON string into slice of records
			var records []utils.Record
			err = json.Unmarshal([]byte(recordsJSON), &records)
			if err != nil {
				return nil, err
			}

			log.Printf("Found %d records", len(records))
			resp := &RecordListResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register POST /profiles/query
		huma.Register(api, huma.Operation{
			OperationID:   "post-profile-query",
			Method:        http.MethodPost,
			Path:          "/profiles/query",
			Summary:       "Query profiles with selector",
			Tags:          []string{"Profiles"},
			Description:   "Query profiles with a selector. This is the same api as the records/query api.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *ProfileSelector) (*RecordListResponse, error) {
			// Convert selector map to JSON string
			selectorJSON, err := json.Marshal(map[string]interface{}{"selector": input.Body.Selector})
			if err != nil {
				return nil, err
			}

			recordsJSON, err := utils.GetRecordWithSelector(string(selectorJSON), options.LocalTest)
			if err != nil {
				return nil, err
			}

			var records []utils.Record

			err = json.Unmarshal([]byte(recordsJSON), &records)
			if err != nil {
				return nil, err
			}

			resp := &RecordListResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register POST /profiles/update
		huma.Register(api, huma.Operation{
			OperationID:   "post-profile-update",
			Method:        http.MethodPost,
			Path:          "/profiles/update",
			Summary:       "Update a profile",
			Tags:          []string{"Profiles"},
			Description:   "Update a profile. The RecordID is required to identify the profile. The updateInfo is a map of the fields to update (droneID, pilotID, zoneID, reserved), if a field is not provided, that field will not be updated. This is the same api as the records/update api, with the recordType set to 'profile' at the server side.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *ProfileUpdateInput) (*ReceviedRecordResponse, error) {
			// get the fields from the updateInfo map
			droneID, err := utils.GetFieldFromMap(input.Body.UpdateInfo, "droneID")
			if err != nil {
				return nil, err
			}
			pilotID, err := utils.GetFieldFromMap(input.Body.UpdateInfo, "pilotID")
			if err != nil {
				return nil, err
			}
			zoneID, err := utils.GetFieldFromMap(input.Body.UpdateInfo, "zoneID")
			if err != nil {
				return nil, err
			}
			reserved, err := utils.GetFieldFromMap(input.Body.UpdateInfo, "reserved")
			if err != nil {
				return nil, err
			}

			err = utils.UpdateRecord(input.Body.RecordID, droneID, pilotID, zoneID, "profile", reserved, options.LocalTest)
			resp := &ReceviedRecordResponse{}
			resp.Body.Message = "Profile updated successfully."
			resp.Body.Record = utils.Record{
				RecordID:   input.Body.RecordID,
				DroneID:    droneID,
				PilotID:    pilotID,
				ZoneID:     zoneID,
				RecordType: "profile",
				Reserved:   reserved,
			}
			return resp, nil
		})

		// Register POST /records/query
		huma.Register(api, huma.Operation{
			OperationID:   "post-record-query",
			Method:        http.MethodPost,
			Path:          "/records/query",
			Summary:       "Query records with selector",
			Tags:          []string{"Records"},
			Description:   "Query records with a selector.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *RecordSelector) (*RecordListResponse, error) {
			// Convert selector map to JSON string
			selectorJSON, err := json.Marshal(map[string]interface{}{"selector": input.Body.Selector})
			if err != nil {
				return nil, err
			}

			recordsJSON, err := utils.GetRecordWithSelector(string(selectorJSON), options.LocalTest)
			if err != nil {
				return nil, err
			}

			// Parse JSON string into slice of records
			var records []utils.Record
			err = json.Unmarshal([]byte(recordsJSON), &records)
			if err != nil {
				return nil, err
			}

			resp := &RecordListResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register GET /records/all
		huma.Register(api, huma.Operation{
			OperationID:   "get-all-records",
			Method:        http.MethodGet,
			Path:          "/records/all",
			Summary:       "Get all records",
			Tags:          []string{"Records"},
			Description:   "Get all records.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *EmptyInput) (*RecordListResponse, error) {
			recordsJSON, err := utils.GetAllRecords(options.LocalTest)
			if err != nil {
				return nil, err
			}

			var records []utils.Record
			err = json.Unmarshal([]byte(recordsJSON), &records)
			if err != nil {
				return nil, err
			}

			resp := &RecordListResponse{}
			resp.Body.Message = fmt.Sprintf("Found %d records", len(records))
			resp.Body.Records = records
			return resp, nil
		})

		// Register POST /records/update
		huma.Register(api, huma.Operation{
			OperationID:   "post-record-update",
			Method:        http.MethodPost,
			Path:          "/records/update",
			Summary:       "Update a record",
			Tags:          []string{"Records"},
			Description:   "Update a record.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *RecordInput) (*ReceviedRecordResponse, error) {
			err := utils.UpdateRecord(input.Body.RecordID, input.Body.DroneID, input.Body.PilotID, input.Body.ZoneID, input.Body.RecordType, input.Body.Reserved, options.LocalTest)
			if err != nil {
				return nil, err
			}
			resp := &ReceviedRecordResponse{}
			resp.Body.Message = "Record updated successfully."
			resp.Body.Record = utils.Record{
				RecordID:   input.Body.RecordID,
				DroneID:    input.Body.DroneID,
				PilotID:    input.Body.PilotID,
				ZoneID:     input.Body.ZoneID,
				RecordType: input.Body.RecordType,
				Reserved:   input.Body.Reserved,
			}
			return resp, nil
		})

		// Register POST /records/history
		huma.Register(api, huma.Operation{
			OperationID:   "post-record-history",
			Method:        http.MethodPost,
			Path:          "/records/history",
			Summary:       "Get the history of a record",
			Tags:          []string{"Records"},
			Description:   "Get the history of a record. This api requires the recordID to be provided.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *RecordIDInput) (*HistoryQueryResultList, error) {
			historyJSON, err := utils.GetRecordHistory(input.Body.RecordID, options.LocalTest)
			if err != nil {
				return nil, err
			}

			var history []HistoryQueryResult
			err = json.Unmarshal([]byte(historyJSON), &history)
			if err != nil {
				return nil, err
			}

			resp := &HistoryQueryResultList{}
			resp.Body.Message = fmt.Sprintf("Found %d history records", len(history))
			resp.Body.History = history
			return resp, nil
		})


		// ==================== Hash API endpoints ====================

		// Register POST /hash/store (authenticated — write operation)
		huma.Register(authAPI, huma.Operation{
			OperationID:   "post-hash-store",
			Method:        http.MethodPost,
			Path:          "/hash/store",
			Summary:       "Store a hash",
			Tags:          []string{"Hash"},
			Description:   "Store an (identifier, hash) pair on the blockchain.",
			DefaultStatus: http.StatusCreated,
			Security:      []map[string][]string{{"bearerAuth": {}}},
		}, func(ctx context.Context, input *HashStoreInput) (*HashStoreResponse, error) {
			if input.Body.Identifier == "" {
				return nil, huma.Error400BadRequest("identifier is required")
			}
			if input.Body.Hash == "" {
				return nil, huma.Error400BadRequest("hash is required")
			}

			// Use CreateRecord: identifier→recordID, hash→reserved, recordType="hash"
			err := utils.CreateRecord(
				input.Body.Identifier, // recordID
				"",                    // droneID (unused)
				"",                    // pilotID (unused)
				"",                    // zoneID (unused)
				"hash",               // recordType
				input.Body.Hash,      // reserved = hash value
				options.LocalTest,
			)
			if err != nil {
				detail := &huma.ErrorDetail{
					Message:  fmt.Sprintf("identifier '%s' already exists", input.Body.Identifier),
					Location: "body.identifier",
				}
				if existingJSON, readErr := utils.ReadRecord(input.Body.Identifier, options.LocalTest); readErr == nil {
					var existing utils.Record
					if json.Unmarshal([]byte(existingJSON), &existing) == nil {
						detail.Value = map[string]string{
							"identifier": existing.RecordID,
							"hash":       existing.Reserved,
						}
					}
				}
				return nil, huma.Error409Conflict(
					fmt.Sprintf("failed to store hash: %v", err),
					detail,
				)
			}

			resp := &HashStoreResponse{}
			resp.Body.Message = "Hash stored successfully."
			return resp, nil
		})

		// Register POST /hash/retrieve (no auth — read operation)
		huma.Register(api, huma.Operation{
			OperationID:   "post-hash-retrieve",
			Method:        http.MethodPost,
			Path:          "/hash/retrieve",
			Summary:       "Retrieve a hash",
			Tags:          []string{"Hash"},
			Description:   "Retrieve a stored hash by its identifier.",
			DefaultStatus: http.StatusOK,
		}, func(ctx context.Context, input *HashRetrieveInput) (*HashRetrieveResponse, error) {
			if input.Body.Identifier == "" {
				return nil, huma.Error400BadRequest("identifier is required")
			}

			recordJSON, err := utils.ReadRecord(input.Body.Identifier, options.LocalTest)
			if err != nil {
				return nil, huma.Error404NotFound(
					fmt.Sprintf("hash not found for identifier '%s': %v", input.Body.Identifier, err),
				)
			}

			var record utils.Record
			if err := json.Unmarshal([]byte(recordJSON), &record); err != nil {
				return nil, err
			}

			resp := &HashRetrieveResponse{}
			resp.Body.Identifier = record.RecordID
			resp.Body.Hash = record.Reserved
			return resp, nil
		})


		
		hooks.OnStart(func() {
			if !options.LocalTest {
				defer utils.ClientConn.Close()
				defer utils.GatewayConn.Close()
			}
			fmt.Printf("Starting server on port %d...\n", options.Port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router); err != nil {
   			 log.Fatalf("Server failed: %v", err)
				}
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}

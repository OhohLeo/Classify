package core

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/foize/go.fifo"
	"github.com/hydrogen18/stoppableListener"
	"net"
	"net/http"
)

type Server struct {
	api       *rest.Api
	config    *ServerConfig
	stoppable *stoppableListener.StoppableListener
	events    *fifo.Queue
}

type ServerConfig struct {
	Url string `json:"url"`
}

type ProtocolReq struct {
	Collection string
	Type       string
	Data       interface{}
}

type Event struct {
	Event  string      `json:"event"`
	Status string      `json:"status"`
	Name   string      `json:"name"`
	Data   interface{} `json:"data"`
}

// ServerStart launches web server
func (c *Classify) CreateServer(config ServerConfig) (server *Server, err error) {

	if config.Url == "" {
		err = fmt.Errorf("No server configuration found!")
		return
	}

	server = new(Server)

	// Stockage de la configuration
	server.config = &config

	// Init events channel
	server.events = fifo.NewQueue()

	listener, err := net.Listen("tcp", config.Url)
	if err != nil {
		return
	}

	server.stoppable, err = stoppableListener.New(listener)
	if err != nil {
		return
	}

	api := rest.NewApi()

	api.Use(rest.DefaultDevStack...)

	// Enable CORS
	api.Use(&rest.CorsMiddleware{
		RejectNonCorsRequests: false,
		OriginValidator: func(origin string, request *rest.Request) bool {
			fmt.Printf("REQUEST %s %+v\n", origin, request.URL)
			return true
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders: []string{
			"Accept", "Content-Type", "Origin"},
		AccessControlAllowCredentials: true,
		AccessControlMaxAge:           3600,
	})

	router, err := rest.MakeRouter(

		// Establish connection to the web-services
		rest.Get("/stream", server.HandleStream),

		// Handle references
		rest.Get("/references", c.ApiGetReferences),

		// Handle imports
		rest.Post("/imports", c.ApiAddImport),
		rest.Get("/imports", c.ApiGetImports),
		rest.Delete("/imports", c.ApiDeleteImport),
		rest.Put("/imports/start", c.ApiStartImport),
		rest.Put("/imports/stop", c.ApiStopImport),
		rest.Get("/imports/:name/config", c.ApiGetImportConfig),
		rest.Patch("/imports/:name/config", c.ApiPatchImportConfig),
		rest.Put("/imports/:name/:param", c.ApiPutImportParam),

		// Handle exports
		rest.Post("/exports", c.ApiAddExport),
		rest.Get("/exports", c.ApiGetExports),
		rest.Delete("/exports", c.ApiDeleteExport),
		rest.Put("/exports/force", c.ApiForceExport),
		rest.Put("/exports/stop", c.ApiStopExport),
		rest.Get("/exports/:name/config", c.ApiGetExportConfig),
		rest.Patch("/exports/:name/config", c.ApiPatchExportConfig),
		rest.Put("/exports/:name/:param", c.ApiPutExportParam),

		// Handle collections
		rest.Post("/collections", c.ApiPostCollection),
		rest.Get("/collections", c.ApiGetCollections),
		rest.Get("/collections/:name", c.ApiGetCollectionByName),
		rest.Patch("/collections/:name", c.ApiPatchCollection),
		rest.Delete("/collections/:name", c.ApiDeleteCollectionByName),
		rest.Get("/collections/:name/config", c.ApiGetCollectionConfig),
		rest.Patch("/collections/:name/config", c.ApiPatchCollectionConfig),

		// Handle collection buffer
		rest.Get("/collections/:name/buffers",
			c.ApiGetCollectionBuffers),
		rest.Delete("/collections/:name/buffers",
			c.ApiDeleteCollectionBuffers),
		rest.Get("/collections/:name/buffers/:id",
			c.ApiGetCollectionSingleBuffer),
		rest.Patch("/collections/:name/buffers/:id",
			c.ApiPatchCollectionSingleBuffer),
		rest.Delete("/collections/:name/buffers/:id",
			c.ApiDeleteCollectionSingleBuffer),
		rest.Post("/collections/:name/buffers/:id/validate",
			c.ApiValidateCollectionSingleBuffer),

		// Handle collection items
		rest.Get("/collections/:name/items",
			c.ApiGetCollectionItems),
		rest.Delete("/collections/:name/items",
			c.ApiDeleteCollectionItems),

		rest.Get("/collections/:name/items/:id",
			c.ApiGetCollectionSingleItem),
		rest.Patch("/collections/:name/items/:id",
			c.ApiPatchCollectionSingleItem),
		rest.Delete("/collections/:name/items/:id",
			c.ApiDeleteCollectionSingleItem),
	)

	if err != nil {
		return
	}

	api.SetApp(router)

	// Store api
	server.api = api

	return
}

func (c *Classify) SendEvent(event string, status string, name string, data interface{}) {
	c.Server.SendEvent(event, status, name, data)
}

func (s *Server) Start() {

	http.Handle("/", http.FileServer(http.Dir("www")))

	log.Println("Serving at " + s.config.Url)
	http.Serve(s.stoppable, s.api.MakeHandler())
}

// ServerStop stop web server
func (s *Server) Stop() {
	log.Println("Stop server")
	s.stoppable.Stop()
}

// SendEvent add new event on the event channel
func (s *Server) SendEvent(eventType string, status string, name string, data interface{}) {

	fmt.Printf("SEND EVENT %s [%s] name:%s %+v\n", eventType, status, name, data)

	s.events.Add(Event{
		Event:  eventType,
		Status: status,
		Name:   name,
		Data:   data,
	})
}

var idx int = 0

func (s *Server) HandleStream(w rest.ResponseWriter, r *rest.Request) {

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		rest.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Prepare write response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Close notification
	notify := w.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case <-notify:
			return
		default:
			event, ok := s.events.Next().(Event)
			if ok {

				eventJson, err := json.Marshal(event)
				if err != nil {
					rest.Error(w, "Encoding event error", http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(w.(http.ResponseWriter), "data: %s\n\n", eventJson)

				// Send data immediately
				flusher.Flush()
			}
		}
	}
}

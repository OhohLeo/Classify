package core

import (
	"encoding/json"
	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
)

// getIdsAndCollections get from Url parameters imports and the collections
func (c *Classify) getIdsAndCollections(r *rest.Request) (imports map[string]Import, collections map[string]Collection, err error) {

	// From the url query list
	values := r.URL.Query()

	// Check and get the import list
	imports, err = c.GetImportsByIds(values["id"])
	if err != nil {
		return
	}

	// Check and get the collection list
	collections, err = c.GetCollectionsByNames(values["collection"])
	if err != nil {
		return
	}

	return
}

type ApiAddImportsBody struct {
	Type        string          `json:"type"`
	Collections []string        `json:"collections"`
	Params      json.RawMessage `json:"params"`
}

// PostCollectionImport add a new import to the collection specified
// POST /imports
func (c *Classify) ApiAddImport(w rest.ResponseWriter, r *rest.Request) {

	// Get import parameters
	var body ApiAddImportsBody
	err := r.DecodeJsonPayload(&body)
	if err != nil {
		rest.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	// Check and get the collection list
	collections, err := c.GetCollectionsByNames(body.Collections)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	i, err := c.AddImport(body.Type, body.Params, collections)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("References", i.Id)

	w.WriteHeader(http.StatusNoContent)
}

// List all the imports selected by id or by collections
// GET /imports?id=IMPORT_ID&collection=COLLECTION_NAME
func (c *Classify) ApiGetImports(w rest.ResponseWriter, r *rest.Request) {

	ids, collections, err := c.getIdsAndCollections(r)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := c.GetImports(ids, collections)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteJson(res)
}

// ApiDeleteImport remove specified import selected by id and by the
// collections
// DELETE /imports?id=IMPORT_ID&collection=COLLECTION_NAME
func (c *Classify) ApiDeleteImport(w rest.ResponseWriter, r *rest.Request) {

	ids, collections, err := c.getIdsAndCollections(r)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.DeleteImports(ids, collections); err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Launch the analysis of the collection import
// PUT /imports/start?id=IMPORT_ID&collection=COLLECTION_NAME
func (c *Classify) ApiStartImport(w rest.ResponseWriter, r *rest.Request) {

	ids, collections, err := c.getIdsAndCollections(r)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.StartImports(ids, collections); err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stop the analysis of the collection import
// PUT /imports/stop?id=IMPORT_ID&collection=COLLECTION_NAME
func (c *Classify) ApiStopImport(w rest.ResponseWriter, r *rest.Request) {

	ids, collections, err := c.getIdsAndCollections(r)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.StopImports(ids, collections); err != nil {
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

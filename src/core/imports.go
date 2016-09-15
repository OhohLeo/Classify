package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ohohleo/classify/imports"
	"github.com/ohohleo/classify/imports/directory"
)

// Type of imports
var newImports = map[string]func(json.RawMessage) (imports.Import, error){
	"directory": func(input json.RawMessage) (i imports.Import, err error) {
		var directory directory.Directory
		err = json.Unmarshal(input, &directory)
		if err == nil {
			i = &directory
		}
		return
	},
}

type Import struct {
	Id          string
	engine      imports.Import
	collections map[string]Collection
}

func (i *Import) HasCollection(name string) (ok bool) {
	_, ok = i.collections[name]
	return
}

// Return true if import has a specified collection or no collections are specified
func (i *Import) HasCollections(collections map[string]Collection) bool {

	if len(collections) > 0 {

		for name, _ := range collections {
			if i.HasCollection(name) {
				return true
			}
		}

		// No collection match
		return false
	}

	return true
}

// Check imports ids and return the list of imports
func (c *Classify) GetImportsByIds(ids []string) (imports map[string]Import, err error) {

	imports = make(map[string]Import)

	for _, id := range ids {
		i, ok := c.imports[id]
		if ok == false {
			err = fmt.Errorf("import '%s' not existing", id)
			return
		}

		imports[id] = i
	}

	return
}

// Add new import process
func (c *Classify) AddImport(importType string, params json.RawMessage, collections map[string]Collection) (i Import, err error) {

	// Nécessite l'existence d'au moins une collection
	if len(collections) < 1 {
		err = errors.New("required at least one existing collection")
	}

	// Field required
	if importType == "" {
		err = errors.New("type field is mandatory")
		return
	}

	// Check that the type exists
	createImport, ok := newImports[importType]
	if ok == false {
		err = errors.New("import type '" + importType + "' not handled")
		return
	}

	// Create new import
	importEngine, err := createImport(params)
	if err != nil {
		return
	}

	alreadyExists := false

	// Check if similar import already exists
	for _, i = range c.imports {

		// Returns similar import found
		if i.engine.GetType() == importType && i.engine.Eq(importEngine) {
			alreadyExists = true
			break
		}
	}

	// Otherwise create your import structure
	if alreadyExists == false {
		id := getRandomName()
		i = Import{
			Id:          id,
			engine:      importEngine,
			collections: collections,
		}

		if c.imports == nil {
			c.imports = make(map[string]Import)
		}

		// Store the new import
		c.imports[id] = i
		return
	}

	i.collections = collections
	return
}

// Remove import from the list
func (c *Classify) DeleteImports(ids map[string]Import, collections map[string]Collection) (err error) {

	// At least one import id or one collection must be specified
	if len(ids) == 0 && len(collections) == 0 {
		err = errors.New("required import ids or collection names")
		return
	}

	// If no ids are specified : remove all import relative to the
	// same collection
	if len(ids) == 0 {
		ids = c.imports
	}

	for id, i := range ids {

		// Unlink the collection with the specified import
		for name, _ := range collections {
			delete(i.collections, name)
		}

		// If no collection are linked with specified import
		if len(i.collections) < 1 {

			// Remove the import
			delete(c.imports, id)
		}
	}
	return
}

// Get the whole list of imports by Type
func (c *Classify) GetImports(ids map[string]Import, collections map[string]Collection) (res map[string]map[string]imports.Import, err error) {

	res = make(map[string]map[string]imports.Import)

	// If no ids are specified : get all
	if len(ids) == 0 {
		ids = c.imports
	}

	for name, i := range ids {

		if i.HasCollections(collections) == false {
			continue
		}

		t := i.engine.GetType()

		if res[t] == nil {
			res[t] = make(map[string]imports.Import)
		}

		res[t][name] = i.engine
	}

	return
}

// Launch the process of importation of specified import
func (c *Classify) StartImports(ids map[string]Import, collections map[string]Collection) error {

	// If no ids are specified : get all
	if len(ids) == 0 {
		ids = c.imports
	}

	// Get the import channel
	for _, i := range ids {

		if i.HasCollections(collections) == false {
			continue
		}

		channel, err := i.engine.Start()
		if err != nil {
			return err
		}

		// Send all data imported to the collections
		go func() {
			for {
				if input, ok := <-channel; ok {

					// For each collections linked with the importation
					for _, collection := range i.collections {

						// Distribute the new value
						collection.OnInput(input)
					}
					continue
				}
				break
			}
		}()
	}

	return nil
}

// Stop the importing process
func (c *Classify) StopImports(ids map[string]Import, collections map[string]Collection) error {

	// If no ids are specified : get all
	if len(ids) == 0 {
		ids = c.imports
	}

	for id, i := range ids {

		if i.HasCollections(collections) == false {
			continue
		}

		c.imports[id].engine.Stop()
	}

	return nil
}
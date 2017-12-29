package core

import (
	"github.com/ohohleo/classify/imports"
	"github.com/ohohleo/classify/requests"
	"github.com/stretchr/testify/assert"
	//"golang.org/x/net/websocket"
	//"log"
	"testing"
)

const URL = "http://127.0.0.1:3333"

func TestApi(t *testing.T) {
	assert := assert.New(t)

	requests.New(2, false)

	var classify *Classify

	go func() {
		checkGetReferences(assert)
		checkPostCollection(assert)
		checkGetCollections(assert)
		checkGetCollectionByName(assert)
		checkAddImport(assert)
		checkGetImports(assert)
		// checkStartImport(assert)
		// checkStopImport(assert)
		//checkDeleteImport(assert)
		checkDeleteCollection(assert)
		classify.Stop()
	}()

	var err error
	classify, err = Start()
	assert.Nil(err)

	// Launch server
	classify.Server.Start()
}

func checkGetReferences(assert *assert.Assertions) {
	var rsp GetReferences
	c, err := requests.Send("GET", URL+"/references", nil, nil, &rsp)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(200, result.Status)

	assert.Equal(GetReferences{
		Websites: []string{"IMDB"},
		Types:    []string{"movies"},
	}, rsp)
}

func checkPostCollection(assert *assert.Assertions) {
	var rsp map[string]string

	// Failure : collection type doesn't exist
	c, err := requests.Send("POST", URL+"/collections",
		map[string]string{
			"Content-Type": "application/json",
		},
		ApiCollectionBody{
			Name: "test",
			Type: "error",
		}, &rsp)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "invalid collection type 'error'",
	}, rsp)

	// Success : collection created
	c, err = requests.Send("POST", URL+"/collections",
		map[string]string{
			"Content-Type": "application/json",
		},
		ApiCollectionBody{
			Name: "test",
			Type: "movies",
		}, nil)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(204, result.Status)

	// Failure : collection already created
	c, err = requests.Send("POST", URL+"/collections",
		map[string]string{
			"Content-Type": "application/json",
		},
		ApiCollectionBody{
			Name: "test",
			Type: "movies",
		}, &rsp)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "collection 'test' already exists",
	}, rsp)
}

func checkGetCollections(assert *assert.Assertions) {

	var rsp []ApiCollection

	// Success : get collections list
	c, err := requests.Send("GET", URL+"/collections",
		nil, nil, &rsp)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(200, result.Status)

	assert.Equal([]ApiCollection{
		ApiCollection{
			Name: "test",
			Type: "movies",
		},
	}, rsp)
}

func checkGetCollectionByName(assert *assert.Assertions) {

	var rsp Collection

	// Success : get collection 'test'
	c, err := requests.Send("GET", URL+"/collections/test",
		nil, nil, &rsp)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(200, result.Status)

	// TODO get result

	var rspError map[string]string

	// Failure : collection 'test' doesn't exist
	c, err = requests.Send("GET", URL+"/collections/error",
		nil, nil, &rspError)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "collection 'error' not existing",
	}, rspError)
}

func checkAddImport(assert *assert.Assertions) {

	// Success : staet specified collection
	c, err := requests.Send("POST", URL+"/imports",
		map[string]string{
			"Content-Type": "application/json",
		},
		map[string]interface{}{
			"type":        "directory",
			"collections": []string{"test"},
			"params": map[string]interface{}{
				"path":         "/tmp",
				"is_recursive": false,
			},
		}, nil)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(204, result.Status)

	// Failure : the collection doesn't exist
	var rsp map[string]string

	c, err = requests.Send("POST", URL+"/imports",
		map[string]string{
			"Content-Type": "application/json",
		},
		map[string]interface{}{
			"type":        "directory",
			"collections": []string{"error"},
			"params": map[string]interface{}{
				"path":         "/tmp",
				"is_recursive": false,
			},
		},
		&rsp)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "collection 'error' not existing",
	}, rsp)

	// Failure : the import type is not defined
	c, err = requests.Send("POST", URL+"/imports",
		map[string]string{
			"Content-Type": "application/json",
		},
		map[string]interface{}{
			"name":        "ok",
			"type":        "error",
			"collections": []string{"test"},
			"params": map[string]interface{}{
				"path":         "/tmp",
				"is_recursive": false,
			},
		}, &rsp)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "import type 'error' not handled",
	}, rsp)
}

func checkGetImports(assert *assert.Assertions) {

	var rspOk map[string]map[string]imports.Import

	// Success : get collections list
	c, err := requests.Send("GET", URL+"/imports",
		nil, nil, &rspOk)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(200, result.Status)

	assert.Equal(1, len(rspOk))
	_, ok = rspOk["directory"]
	assert.True(ok)

	// Failure : the collection doesn't exist
	var rsp map[string]string

	c, err = requests.Send("GET", URL+"/imports?collection=error",
		nil, nil, &rsp)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "collection 'error' not existing",
	}, rsp)

}

// var ws *websocket.Conn

// func checkStartImport(assert *assert.Assertions) {

// 	var err error

// 	// Establish a web socket connection
// 	ws, err = websocket.Dial(
// 		"ws://localhost:3333/ws", "", "http://localhost/")
// 	assert.Nil(err)

// 	// Receive data from the web socket
// 	go func() {
// 		var msg = make([]byte, 512)
// 		for {
// 			n, err := ws.Read(msg)
// 			if n == 0 {
// 				continue
// 			}

// 			if err != nil {
// 				log.Printf("Error: %s\n", err.Error())
// 			}

// 			log.Printf("Received: %d %s\n", n, msg[:n])
// 		}
// 	}()

// 	// Success : state specified collection
// 	c, err := requests.Send("PUT", URL+"/import//start",
// 		nil, nil, nil)
// 	assert.Nil(err)

// 	result, ok := <-c
// 	assert.True(ok)
// 	assert.Equal(204, result.Status)
// }

// func checkStopImport(assert *assert.Assertions) {

// 	// Success : stop specified collection
// 	c, err := requests.Send("PUT", URL+"/import//stop",
// 		nil, nil, nil)
// 	assert.Nil(err)

// 	result, ok := <-c
// 	assert.True(ok)
// 	assert.Equal(204, result.Status)
// }

func checkPatchCollection(assert *assert.Assertions) {

	// Success : patch collection 'test'
	c, err := requests.Send("PATCH", URL+"/collections/test",
		nil, nil, nil)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(204, result.Status)
}

func checkDeleteCollection(assert *assert.Assertions) {

	// Success : delete specified collection
	c, err := requests.Send("DELETE", URL+"/collections/test",
		nil, nil, nil)
	assert.Nil(err)

	result, ok := <-c
	assert.True(ok)
	assert.Equal(204, result.Status)

	// Failure : the collection doesn't exist
	var rsp map[string]string

	c, err = requests.Send("DELETE", URL+"/collections/test",
		nil, nil, &rsp)
	assert.Nil(err)

	result, ok = <-c
	assert.True(ok)
	assert.Equal(400, result.Status)

	assert.Equal(map[string]string{
		"Error": "collection 'test' not existing",
	}, rsp)
}

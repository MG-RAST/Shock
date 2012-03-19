package main

import (
	"net/http"
	"fmt"
	"goweb"
	"os"
	"strconv"
	e "shock/errors"
	user "shock/user"
	ds "shock/datastore"
	bson "launchpad.net/mgo/bson"
)

type NodeController struct{}

// POST: /node
func (cr *NodeController) Create(cx *goweb.Context) {	
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request); if err != nil {
		// No Auth is not damning. Other errors are probably a dead db connection
		if err.Error() != e.NoAuth {
			if err.Error() == e.MongoDocNotFound {
				cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest) 	
				return
			} else {
				fmt.Println("Error at Auth:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError) 
				return	
			}
		}		
	}
	
	// Fake public user 
	if u == nil {
		u = &user.User{Uuid:""}
	}
	
	// Parse uploaded form 
	params, files, err := ParseMultipartForm(cx.Request)
	if err != nil {
		// If not multipart/form-data it will create an empty node. 
		// TODO: create another request parser for non-multipart request
		// to handle this cleaner.		
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			node, err := ds.CreateNodeUpload(u, params, files); if err != nil {
				fmt.Println("Error at create empty:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError) 
				return
			}
			if node == nil {
				// Not sure how you could get an empty node with no error
				// Assume it's the user's fault
				cx.RespondWithError(http.StatusBadRequest)
				return
			} else {
				cx.RespondWithData(node)
				return
			}
		} else {
			// Some error other than request encoding. Theoretically 
			// could be a lost db connection between user lookup and parsing.
			// Blame the user, Its probaby their fault anyway.
			fmt.Println("Error at empty create:", err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}
	}
	// Create node	
	node, err := ds.CreateNodeUpload(u, params, files)	
	if err != nil {
		fmt.Println("err", err.Error())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}
	cx.RespondWithData(node)
	return
}

// DELETE: /node/{id}
func (cr *NodeController) Delete(id string, cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithError(http.StatusNotImplemented) 
}

// DELETE: /node
func (cr *NodeController) DeleteMany(cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithError(http.StatusNotImplemented) 
}

// GET: /node/{id}
func (cr *NodeController) Read(id string, cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request); if err != nil {
		// No Auth is not damning. Other errors are probably a dead db connection
		if err.Error() != e.NoAuth {
			if err.Error() == e.MongoDocNotFound {
				cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest) 	
				return
			} else {
				fmt.Println("Error at Auth:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError) 
				return	
			}
		}		
	}
	
	// Fake public user 
	if u == nil {
		u = &user.User{Uuid:""}
	}

	// Gather query params and setup flags
	query := cx.Request.URL.Query()
	_, download := query["download"]
	_, pipe := query["pipe"]
	_, list := query["list"]
		
	// Load node and handle user unauthorized
	node, err := ds.LoadNode(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			fmt.Println("Unauthorized")
			cx.RespondWithError(http.StatusUnauthorized) 
			return	
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			fmt.Println("Err@node_Read:LoadNode:", err.Error())		
			cx.RespondWithError(http.StatusInternalServerError) 
			return
		}
	}
	
	// Switch though param flags
	if download {
		nf, err := os.Open(node.DataPath()); if err != nil {
			// File not found or some sort of file read error. 
			// Probably deserves more checking
			fmt.Println("err", err.Error())
			cx.RespondWithError(http.StatusBadRequest) 
			return
		}
		s := &streamer{rs : nf, ws : cx.ResponseWriter, contentType : "application/octet-stream", filename : node.Id, size : node.File.Size}
		err = s.stream()
		if err != nil { fmt.Println("err", err.Error()) }
		return
	} else if pipe {
		cx.RespondWithError(http.StatusNotImplemented) 
	} else if list {
		cx.RespondWithError(http.StatusNotImplemented) 
	} else {
		// Base case respond with node in json	
		cx.RespondWithData(node)
	}
}

// GET: /node
// To do:
// - Iterate node queries
func (cr *NodeController) ReadMany(cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request); if err != nil {
		// No Auth is not damning. Other errors are probably a dead db connection
		if err.Error() != e.NoAuth {
			if err.Error() == e.MongoDocNotFound {
				cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest) 	
				return
			} else {
				fmt.Println("Error at Auth:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError) 
				return	
			}
		}		
	}
	
	// Gather query params and setup flags	
	query := cx.Request.URL.Query()
	l, hasLimit := query["limit"]
	o, hasOffset := query["skip"]
	_, hasQuery := query["query"]

	// Setup query and nodes objects
	q := bson.M{}
	nodes := new(ds.Nodes)

	if u != nil {
		// Admin sees all
		if !u.Admin {		
			q["$or"] = []bson.M{bson.M{"acl.read": []string{}}, bson.M{"acl.read": u.Uuid}}
		}
	} else {
		// select on only nodes with no read rights set
		q["acl.read"] = []string{}
	}
	
	// Gather params to make db query. Do not include the
	// following list.	
	skip := map[string]int{"limit" : 1,"skip" : 1,"query" : 1}
	if hasQuery {
		for key, val := range query {
			_, s := skip[key]
			if !s { 
				q[fmt.Sprintf("attributes.%s",key)] = val[0]
			}
		}
	}
	
	// Limit and skip. Set default if both are not specified
	if hasLimit || hasOffset {
		var lim, off int
		if !hasLimit {
			lim = 100
		} else {
			lim, _ = strconv.Atoi(l[0])
		}
		if !hasOffset {
			off = 0
		} else {
			off, _ = strconv.Atoi(o[0])
		}
		// Get nodes from db
		err := nodes.GetAllLimitOffset(q, lim, off)
		if err != nil {
			fmt.Println("err", err.Error())
			cx.RespondWithError(http.StatusBadRequest) 
			return
		}
	} else {
		// Get nodes from db
		err := nodes.GetAll(q)
		if err != nil {
			fmt.Println("err", err.Error())
			cx.RespondWithError(http.StatusBadRequest) 
			return
		}
	}
	
	cx.RespondWithData(nodes)		
	return
}

// PUT: /node/{id} -> multipart-form 
func (cr *NodeController) Update(id string, cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request); if err != nil {
		// No Auth is not damning. Other errors are probably a dead db connection
		if err.Error() != e.NoAuth {
			if err.Error() == e.MongoDocNotFound {
				cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest) 	
				return
			} else {
				fmt.Println("Error at Auth:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError) 
				return	
			}
		}		
	}
	
	// Fake public user 
	if u == nil {
		u = &user.User{Uuid:""}
	}

	node, err := ds.LoadNode(id, u.Uuid); if err != nil {
		if err.Error() == e.UnAuth {
			fmt.Println("Unauthorized")
			cx.RespondWithError(http.StatusUnauthorized) 
			return	
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			fmt.Println("Err@node_Update:LoadNode:", err.Error())		
			cx.RespondWithError(http.StatusInternalServerError) 
			return
		}		
	}

	params, files, err := ParseMultipartForm(cx.Request)
	if err != nil {
		fmt.Println("err", err.Error())
		cx.RespondWithError(http.StatusBadRequest) 
		return
	} 	

	err = node.Update(params, files)
	if err != nil {
		errors := []string{"node file already set and is immutable", "node file immutable", "node attributes immutable", "node part already exists and is immutable"}
		for e := range errors {
			if err.Error() == errors[e] {
				cx.RespondWithErrorMessage(err.Error(),http.StatusBadRequest)	
				return
			}
		}
		fmt.Println("err", err.Error())
		cx.RespondWithError(http.StatusBadRequest) 
		return
	}	
	cx.RespondWithData(node)
	return	
}

// PUT: /node
func (cr *NodeController) UpdateMany(cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithError(http.StatusNotImplemented) 
}

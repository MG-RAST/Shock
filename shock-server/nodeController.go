package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/filter"
	"github.com/MG-RAST/Shock/store/indexer"
	"github.com/MG-RAST/Shock/store/user"
	"github.com/jaredwilkening/goweb"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"strconv"
)

type NodeController struct{}

func handleAuthError(err error, cx *goweb.Context) {
	switch err.Error() {
	case e.MongoDocNotFound:
		cx.RespondWithErrorMessage("Invalid username or password", http.StatusBadRequest)
		return
	case e.InvalidAuth:
		cx.RespondWithErrorMessage("Invalid Authorization header", http.StatusBadRequest)
		return
	}
	fmt.Println("Error at Auth:", err.Error())
	cx.RespondWithError(http.StatusInternalServerError)
	return
}

// POST: /node
func (cr *NodeController) Create(cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// Fake public user 
	if u == nil {
		if conf.ANON_WRITE {
			u = &user.User{Uuid: ""}
		} else {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		}
	}

	// Parse uploaded form 
	params, files, err := ParseMultipartForm(cx.Request)
	if err != nil {
		// If not multipart/form-data it will create an empty node. 
		// TODO: create another request parser for non-multipart request
		// to handle this cleaner.		
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			node, err := store.CreateNodeUpload(u, params, files)
			if err != nil {
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
			fmt.Println("Error parsing form:", err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}
	}
	// Create node	
	node, err := store.CreateNodeUpload(u, params, files)
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
// ToDo: clean up this function. About to get unmanageable
func (cr *NodeController) Read(id string, cx *goweb.Context) {
	// Log Request and check for Auth
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// Fake public user 
	if u == nil {
		if conf.ANON_READ {
			u = &user.User{Uuid: ""}
		} else {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		}
	}

	// Gather query params
	query := &Query{list: cx.Request.URL.Query()}

	var fFunc filter.FilterFunc = nil
	if query.Has("filter") {
		if filter.Has(query.Value("filter")) {
			fFunc = filter.Filter(query.Value("filter"))
		}
	}

	// Load node and handle user unauthorized
	node, err := store.LoadNode(id, u.Uuid)
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
	// ?download=1
	if query.Has("download") {
		if !node.HasFile() {
			cx.RespondWithErrorMessage("node file not found", http.StatusBadRequest)
			return
		}

		//_, chunksize := 
		// ?index=foo
		if query.Has("index") {
			// if forgot ?part=N
			if !query.Has("part") {
				cx.RespondWithErrorMessage("Index parameter requires part parameter", http.StatusBadRequest)
				return
			}
			// open file
			r, err := os.Open(node.FilePath())
			defer r.Close()
			if err != nil {
				fmt.Println("Err@node_Read:Open:", err.Error())
				cx.RespondWithError(http.StatusInternalServerError)
				return
			}
			// load index
			idx, err := node.Index(query.Value("index"))
			if err != nil {
				cx.RespondWithErrorMessage("Invalid index", http.StatusBadRequest)
				return
			}
			if idx.Type() == "virtual" {
				csize := int64(1048576)
				if query.Has("chunksize") {
					csize, err = strconv.ParseInt(query.Value("chunksize"), 10, 64)
					if err != nil {
						cx.RespondWithErrorMessage("Invalid chunksize", http.StatusBadRequest)
						return
					}
				}
				idx.Set(map[string]interface{}{"ChunkSize": csize})
			}
			var size int64 = 0
			s := &streamer{rs: []io.ReadCloser{}, ws: cx.ResponseWriter, contentType: "application/octet-stream", filename: node.Id, filter: fFunc}
			for _, p := range query.List("part") {
				pos, length, err := idx.Part(p)
				if err != nil {
					cx.RespondWithErrorMessage("Invalid index part", http.StatusBadRequest)
					return
				}
				size += length
				s.rs = append(s.rs, NewSectionReaderCloser(r, pos, length))
			}
			s.size = size
			err = s.stream()
			if err != nil {
				// fix
				fmt.Println("err", err.Error())
			}
		} else {
			nf, err := os.Open(node.FilePath())
			if err != nil {
				// File not found or some sort of file read error. 
				// Probably deserves more checking
				fmt.Println("err", err.Error())
				cx.RespondWithError(http.StatusBadRequest)
				return
			}
			s := &streamer{rs: []io.ReadCloser{nf}, ws: cx.ResponseWriter, contentType: "application/octet-stream", filename: node.Id, size: node.File.Size, filter: fFunc}
			err = s.stream()
			if err != nil {
				// fix
				fmt.Println("err", err.Error())
			}
		}
		return
	} else if query.Has("pipe") {
		cx.RespondWithError(http.StatusNotImplemented)
	} else if query.Has("list") {
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
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// Gather query params
	query := &Query{list: cx.Request.URL.Query()}

	// Setup query and nodes objects
	q := bson.M{}
	nodes := new(store.Nodes)

	if u != nil {
		// Admin sees all
		if !u.Admin {
			q["$or"] = []bson.M{bson.M{"acl.read": []string{}}, bson.M{"acl.read": u.Uuid}}
		}
	} else {
		if conf.ANON_READ {
			// select on only nodes with no read rights set
			q["acl.read"] = []string{}
		} else {
			cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
			return
		}
	}

	// Gather params to make db query. Do not include the
	// following list.	
	skip := map[string]int{"limit": 1, "skip": 1, "query": 1}
	if query.Has("query") {
		for key, val := range query.All() {
			_, s := skip[key]
			if !s {
				q[fmt.Sprintf("attributes.%s", key)] = val[0]
			}
		}
	}

	// Limit and skip. Set default if both are not specified
	if query.Has("limit") || query.Has("skip") {
		var lim, off int
		if query.Has("limit") {
			lim, _ = strconv.Atoi(query.Value("limit"))
		} else {
			lim = 100
		}
		if query.Has("skip") {
			off, _ = strconv.Atoi(query.Value("skip"))
		} else {
			off = 0
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
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// Gather query params
	query := &Query{list: cx.Request.URL.Query()}

	// Fake public user 
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	node, err := store.LoadNode(id, u.Uuid)
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
			fmt.Println("Err@node_Update:LoadNode:", err.Error())
			cx.RespondWithError(http.StatusInternalServerError)
			return
		}
	}

	if query.Has("index") {
		if !node.HasFile() {
			cx.RespondWithErrorMessage("node file empty", http.StatusBadRequest)
			return
		}
		newIndexer := indexer.Indexer(query.Value("index"))
		f, _ := os.Open(node.FilePath())
		defer f.Close()
		idxer := newIndexer(f)
		err := idxer.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = idxer.Dump(node.IndexPath() + "/record")
		if err != nil {
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		} else {
			cx.RespondWithOK()
			return
		}
	} else {
		params, files, err := ParseMultipartForm(cx.Request)
		if err != nil {
			fmt.Println("err", err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}

		err = node.Update(params, files)
		if err != nil {
			errors := []string{e.FileImut, e.AttrImut, "parts cannot be less than 1"}
			for e := range errors {
				if err.Error() == errors[e] {
					cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
					return
				}
			}
			fmt.Println("err", err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}
		cx.RespondWithData(node)
	}
	return
}

// PUT: /node
func (cr *NodeController) UpdateMany(cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithError(http.StatusNotImplemented)
}

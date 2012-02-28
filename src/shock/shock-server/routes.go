package main

import (
	"net/http"
	"fmt"
	"goweb"
	"os"
	ds "shock/datastore"
)

// GET
// /node/{id} 
//           ?download[&index={index}[&part={part}]]
//           ?pipe(&{func}={funcOptions})+)
//           ?list={indexes||functions||parts&index={index}...}
// /node
//           ?paginate[&limit={limit}&offset={offset}]
//           ?query={queryString}[&paginate[&limit={limit}&offset={offset}]]

// PUT
// /node/{id}
//           ?pipe(&{func}={funcOptions})+
//           ?index={type}[&options={options}]
//           ?attributes 
//            -> multipart-form or json file as body
//           ?file[&part={part}] 
//            -> multipart-form or data file as body

// POST
// /node
//       multipart-form containing: data file and/or attributes (json file)
//       empty body

type NodeController struct{}

// POST: /node
//            multipart-form containing: data file and/or attributes (json file)
//            empty body
func (cr *NodeController) Create(cx *goweb.Context) {
	fmt.Println("POST: /node")
	params, files, err := ParseMultipartForm(cx.Request)
	if err != nil {
		if err.Error() == "request Content-Type isn't multipart/form-data" {
			node, err := ds.CreateNodeUpload(params, files); if err != nil {
				fmt.Println("err", err.Error())
			}
			if node != nil {
				cx.RespondWithData(node)
				return
			} else {
				cx.RespondWithError(http.StatusBadRequest)
				return
			}
		} 
		fmt.Println("error:", err.Error())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}	
	node, err := ds.CreateNodeUpload(params, files)	
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
	fmt.Printf("DELETE: /node/%s\n", id)	
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"delete operation currently not supported\" }")
}

// DELETE: /node
func (cr *NodeController) DeleteMany(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"deletemany operation currently not supported\" }")
}

// GET: /node/{id}
//                ?download[&index={index}[&part={part}]]
//                ?pipe(&{func}={funcOptions})+)
//                ?list={indexes||functions||parts&index={index}...}
func (cr *NodeController) Read(id string, cx *goweb.Context) {
	fmt.Printf("GET: /node/%s\n", id)	
	query := cx.Request.URL.Query()
	_, download := query["download"]
	_, pipe := query["pipe"]
	_, list := query["list"]
	
	node, _ := ds.LoadNode(id)
	if node != nil {
		if download {
			nf, err := os.Open(node.DataPath()); if err != nil {
				fmt.Println("err", err.Error())
				cx.RespondWithError(http.StatusBadRequest) 
				return
			}
			s := &streamer{rs : nf, ws : cx.ResponseWriter, contentType : "application/octet-stream", filename : node.Id, size : node.File.Size}
			err = s.stream()
			if err != nil { fmt.Println("err", err.Error()) }
			return
		} else if pipe {
			cx.RespondWithNotFound()
		} else if list {
			cx.RespondWithNotFound()
		} else {	
			cx.RespondWithData(node)
		}
	} else {
		cx.RespondWithNotFound()
	}
}

// GET: /node
//           ?paginate[&limit={limit}&offset={offset}]
//           ?query={queryString}[&paginate[&limit={limit}&offset={offset}]]
func (cr *NodeController) ReadMany(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"readmany operation currently not supported\" }")
}

// PUT: /node/{id} -> multipart-form 
//                ?pipe(&{func}={funcOptions})+
//                ?index={type}[&options={options}]
//                ?file[&part={part}] 
func (cr *NodeController) Update(id string, cx *goweb.Context) {
	fmt.Printf("PUT: /node/%s\n", id)	
	
	node, err := ds.LoadNode(id); if err != nil {
		// add node not found message
		cx.RespondWithError(http.StatusBadRequest)
		return
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
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"updatemany operation currently not supported\" }")
}

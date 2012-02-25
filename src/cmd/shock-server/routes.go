package main

import (
	"http"
	"fmt"
	"goweb"
	"rand"
	"os"
	"crypto/md5"
	"crypto/sha1"
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
		if err.String() == "request Content-Type isn't multipart/form-data" {
			node, err := ds.CreateNodeUpload(params, files); if err != nil {
				fmt.Println("err", err.String())
			}
			if node != nil {
				cx.RespondWithData(node)
				return
			} else {
				cx.RespondWithError(http.StatusBadRequest)
				return
			}
		} 
		fmt.Println("error:", err.String())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}	
	node, err := ds.CreateNodeUpload(params, files)	
	if err != nil {
		fmt.Println("err", err.String())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}
	cx.RespondWithData(node)
	return
}

// helper function for create & update
func ParseMultipartForm(r *http.Request) (params map[string]string, files ds.FormFiles, err os.Error){
	params = make(map[string]string)
	files = make(ds.FormFiles)
	md5h := md5.New()
	sha1h := sha1.New()	
	reader, err := r.MultipartReader(); if err != nil { return }
	for {
		part, err := reader.NextPart(); if err != nil { break }
		if part.FileName() == "" {
			buffer := make([]byte, 10240)
			n, err := part.Read(buffer)
			if n == 0 || err != nil { break }
			params[part.FormName()] = fmt.Sprintf("%s", buffer[0:n])
		} else {
			tmpPath := fmt.Sprintf("%s/temp/%d%d", DATAROOT, rand.Int(), rand.Int())
			files[part.FormName()] = ds.FormFile{Name: part.FileName(), Path: tmpPath, Checksum: make(map[string]string)}
			tmpFile, err := os.Create(tmpPath); if err != nil { break }
			for {
				buffer := make([]byte, 10240)
				n, err := part.Read(buffer)
				if n == 0 || err != nil { break }
				tmpFile.Write(buffer[0:n])
				md5h.Write(buffer[0:n])
				sha1h.Write(buffer[0:n]) 
			}
			files[part.FormName()].Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum())		
			files[part.FormName()].Checksum["sha1"] = fmt.Sprintf("%x", sha1h.Sum())
		
			tmpFile.Close()
			md5h.Reset()
			sha1h.Reset()
		}
	}
	if err != nil { return }
	return
}

// DELETE: /node
func (cr *NodeController) Delete(id string, cx *goweb.Context) {
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
			cx.RespondWithNotFound()
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
		fmt.Println("err", err.String())
		cx.RespondWithError(http.StatusBadRequest) 
		return
	} 	

	err = node.Update(params, files)
	if err != nil {
		errors := []string{"node file already set and is immutable", "node file immutable", "node attributes immutable", "node part already exists and is immutable"}
		for e := range errors {
			if err.String() == errors[e] {
				cx.RespondWithErrorMessage(err.String(),http.StatusBadRequest)	
				return
			}
		}
		fmt.Println("err", err.String())
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

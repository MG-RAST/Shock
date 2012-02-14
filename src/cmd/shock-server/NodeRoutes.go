package main

import (
	"fmt"
	"goweb"
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

	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

// DELETE: /node
func (cr *NodeController) Delete(id string, cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

// DELETE: /node
func (cr *NodeController) DeleteMany(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

// GET: /node/{id}
//                ?download[&index={index}[&part={part}]]
//                ?pipe(&{func}={funcOptions})+)
//                ?list={indexes||functions||parts&index={index}...}
func (cr *NodeController) Read(id string, cx *goweb.Context) {
	query := cx.Request.URL.Query()
	for key, arr := range query {
		for i := range arr {
			fmt.Println(key, arr[i])
		}
	}
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"id\" : %s }", id)
}

// GET: /node
//           ?paginate[&limit={limit}&offset={offset}]
//           ?query={queryString}[&paginate[&limit={limit}&offset={offset}]]
func (cr *NodeController) ReadMany(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

// PUT: /node/{id}
//                ?pipe(&{func}={funcOptions})+
//                ?index={type}[&options={options}]
//                ?attributes 
//                 -> multipart-form or json file as body
//                ?file[&part={part}] 
//                 -> multipart-form or data file as body
func (cr *NodeController) Update(id string, cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

// PUT: /node
func (cr *NodeController) UpdateMany(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(cx.ResponseWriter, "{ \"message\" : \"operation currently not supported\" }")
}

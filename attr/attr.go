package main

import (
	"flag"
	"fmt"
	"os"
	"io/ioutil"
	"json"
)

/*
{
	foo : {
		foo : bar,
		foo2 : bar 
	}
	foo3 : bar,
	foo4 : [bas, bas, bas, bas]
 }

node_id | level | index | container | tag  | value 
      1 |     0 |       | m         | foo  | 1
      1 |     0 |       |           | foo3 | bar
      1 |     0 |       | a         | foo4 | 2
      1 |     1 |       |           | foo  | bar
      1 |     1 |       |           | foo2 | bar
      1 |     2 |     0 |           | foo4 | bas
      1 |     2 |     1 |           | foo4 | bas
      1 |     2 |     2 |           | foo4 | bas
      1 |     2 |     3 |           | foo4 | bas
*/

type Attributes struct {
	json    interface{}
	jsonRaw []byte
	rows    []AttrRow
}

type AttrRow struct {
	id        int
	level     int
	index     int
	container string
	tag       string
	value     string
}

func (attr *Attributes) Construct() (err os.Error) {
	if attr.jsonRaw == nil {
		return os.NewError("jsonRaw undefined")
	} else if attr.rows != nil {
		return nil
	}
	if attr.json == nil {
		err = json.Unmarshal(attr.jsonRaw, &attr.json)
		if err != nil {
			return err
		}
	}

	for k, v := range attr.json.(map[string]interface{}) {
		attr._insert(k, v)
	}
	fmt.Println("\n", string(attr.jsonRaw))
	return nil
}

func (attr *Attributes) _insert(key string, value interface{}) (err os.Error) {
	switch vv := value.(type) {
	case string:
		fmt.Println("string =>", key, ":", vv)
	case nil:
		fmt.Println("nil =>", key, ":", vv)
	case bool:
		fmt.Println("bool =>", key, ":", vv)
	case float64:
		fmt.Println("float64 =>", key, ":", vv)
	case []interface{}:
		fmt.Println("array =>", key)
		for i, u := range vv {
			fmt.Println("\t", i, ":", u)
		}
	case interface{}:
		fmt.Println("map =>", key)
		for i, u := range vv.(map[string]interface{}) {
			fmt.Println("\t", i, ":", u)
		}
	default:
		fmt.Println(key, "is of a type I don't know how to handle: ", vv)
	}
	return nil
}

func (attr *Attributes) Deconstruct() (err os.Error) {
	if attr.jsonRaw == nil {
		return os.NewError("rows undefined")
	} else if attr.json != nil {
		return nil
	}
	return nil
}

// Command line options
var (
	filename = flag.String("filename", "", "json file to parse")
	jsStr    interface{}
	err      os.Error
)

func main() {
	flag.Parse()
	fmt.Println("File to parse " + *filename)
	test := new(Attributes)

	test.jsonRaw, err = ioutil.ReadFile(*filename)
	if err != nil {
		fmt.Println("Oh hell..." + err.String())
	}

	test.Construct()
}

package main

import (
	"fmt"
	"goweb"
	conf "shock/conf"
)

func main() {
	goweb.ConfigureDefaultFormatters()
	goweb.MapRest("/node", new(NodeController))
	fmt.Printf("Shock port:%d...starting\n", *conf.PORT)
	goweb.ListenAndServe(":"+fmt.Sprintf("%d", *conf.PORT))	
}
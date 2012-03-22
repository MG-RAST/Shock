package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/goweb"
)

func main() {
	goweb.ConfigureDefaultFormatters()
	goweb.MapRest("/node", new(NodeController))
	goweb.MapRest("/user", new(UserController))
	fmt.Printf("Shock port:%d...starting\n", *conf.PORT)
	goweb.ListenAndServe(":" + fmt.Sprintf("%d", *conf.PORT))
}

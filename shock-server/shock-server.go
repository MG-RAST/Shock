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
	fmt.Printf("Shock (port:%d, dataroot:%q, mongodb_host:%q, secretkey:%q)... starting\n", *conf.PORT, *conf.DATAROOT, *conf.MONGODB, *conf.SECRETKEY)
	goweb.ListenAndServe(":" + fmt.Sprintf("%d", *conf.PORT))
}

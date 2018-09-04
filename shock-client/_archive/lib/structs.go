package lib

import (
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/acl"
)

type Node node.Node

type User struct {
	Username string
	Password string
	Token    string
	Expire   string
}

type Wrapper struct {
	Data  interface{} `json:"data"`
	Error *[]string   `json:"error"`
	//status already parsed in res.Status
}

type WNode struct {
	Status int       `json:"status"`
	Data   *Node     `json:"data"`
	Error  *[]string `json:"error"`
}

type WAcl struct {
	Data  *acl.Acl  `json:"data"`
	Error *[]string `json:"error"`
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	UserName    string `json:"user_name"`
}

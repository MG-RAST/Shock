package lib

import (
	"github.com/MG-RAST/Shock/shock-server/store"
)

type Node store.Node

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
	Data  *Node     `json:"data"`
	Error *[]string `json:"error"`
}

type WAcl struct {
	Data  *store.Acl `json:"data"`
	Error *[]string  `json:"error"`
}

type Token struct {
	AccessToken     string      `json:"access_token"`
	AccessTokenHash string      `json:"access_token_hash"`
	ClientId        string      `json:"client_id"`
	ExpiresIn       int         `json:"expires_in"`
	Expiry          int         `json:"expiry"`
	IssuedOn        int         `json:"issued_on"`
	Lifetime        int         `json:"lifetime"`
	Scopes          interface{} `json:"scopes"`
	TokenId         string      `json:"token_id"`
	TokeType        string      `json:"token_type"`
	UserName        string      `json:"user_name"`
}

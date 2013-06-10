package lib

type User struct {
	Username string
	Password string
	Token    string
	Expire   string
}

type Wrapper struct {
	Data  interface{} `json:"D"`
	Error *[]string   `json:"E"`
	//status already parsed in res.Status
}

type WNode struct {
	Data  *Node     `json:"D"`
	Error *[]string `json:"E"`
}

type WAcl struct {
	Data  *acl      `json:"D"`
	Error *[]string `json:"E"`
}

type Node struct {
	Id           string            `bson:"id" json:"id"`
	Version      string            `bson:"version" json:"version"`
	File         file              `bson:"file" json:"file"`
	Attributes   interface{}       `bson:"attributes" json:"attributes"`
	Indexes      indexes           `bson:"indexes" json:"indexes"`
	Acl          acl               `bson:"acl" json:"-"`
	VersionParts map[string]string `bson:"version_parts" json:"-"`
	Tags         []string          `bson:"tags" json:"tags"`
	Revisions    []Node            `bson:"revisions" json:"-"`
	Linkages     []linkage         `bson:"linkage" json:"linkages"`
}

type file struct {
	Name         string            `bson:"name" json:"name"`
	Size         int64             `bson:"size" json:"size"`
	Checksum     map[string]string `bson:"checksum" json:"checksum"`
	Format       string            `bson:"format" json:"format"`
	Path         string            `bson:"path" json:"-"`
	Virtual      bool              `bson:"virtual" json:"virtual"`
	VirtualParts []string          `bson:"virtual_parts" json:"virtual_parts"`
}

type partsList struct {
	Count  int         `json:"count"`
	Length int         `json:"length"`
	Parts  []partsFile `json:"parts"`
}

type linkage struct {
	Type      string   `bson: "relation" json:"relation"`
	Ids       []string `bson:"ids" json:"ids"`
	Operation string   `bson:"operation" json:"operation"`
}

type indexes map[string]IdxInfo

type IdxInfo struct {
	Type        string `bson:"index_type" json:"-"`
	TotalUnits  int64  `bson:"total_units" json:"total_units"`
	AvgUnitSize int64  `bson:"average_unit_size" json:"average_unit_size"`
}

type acl struct {
	Owner  string   `bson:"owner" json:"owner"`
	Read   []string `bson:"read" json:"read"`
	Write  []string `bson:"write" json:"write"`
	Delete []string `bson:"delete" json:"delete"`
}

type partsFile []string

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

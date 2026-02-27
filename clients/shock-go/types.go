package shock

import "time"

// Node represents a Shock data node.
type Node struct {
	Id           string            `json:"id"`
	Version      string            `json:"version"`
	File         File              `json:"file"`
	Attributes   interface{}       `json:"attributes"`
	Indexes      map[string]*IdxInfo `json:"indexes"`
	VersionParts map[string]string `json:"version_parts"`
	Tags         []string          `json:"tags"`
	Linkages     []Linkage         `json:"linkage"`
	Priority     int               `json:"priority"`
	CreatedOn    time.Time         `json:"created_on"`
	LastModified time.Time         `json:"last_modified"`
	Expiration   time.Time         `json:"expiration"`
	Type         string            `json:"type"`
	Parts        *PartsList        `json:"parts"`
	Locations    []Location        `json:"locations"`
	Restore      bool              `json:"restore"`
}

// File represents file metadata within a node.
type File struct {
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	Checksum     map[string]string `json:"checksum"`
	Format       string            `json:"format"`
	Virtual      bool              `json:"virtual"`
	VirtualParts []string          `json:"virtual_parts"`
	CreatedOn    time.Time         `json:"created_on"`
	Locked       *LockInfo         `json:"locked"`
}

// LockInfo represents a lock on a file or index.
type LockInfo struct {
	CreatedOn time.Time `json:"created_on"`
	Error     string    `json:"error"`
}

// IdxInfo represents index metadata.
type IdxInfo struct {
	TotalUnits  int64     `json:"total_units"`
	AvgUnitSize int64     `json:"average_unit_size"`
	CreatedOn   time.Time `json:"created_on"`
	Locked      *LockInfo `json:"locked"`
}

// Linkage represents a relationship between nodes.
type Linkage struct {
	Type      string   `json:"relation"`
	Ids       []string `json:"ids"`
	Operation string   `json:"operation"`
}

// Location represents a storage location for a node.
type Location struct {
	ID            string     `json:"id"`
	Stored        bool       `json:"stored,omitempty"`
	RequestedDate *time.Time `json:"requestedDate,omitempty"`
}

// PartsList represents the parts metadata of a node.
type PartsList struct {
	Count       int        `json:"count"`
	Length      int        `json:"length"`
	VarLen      bool       `json:"varlen"`
	Parts       [][]string `json:"parts"`
	Compression string     `json:"compression"`
}

// DisplayAcl represents access control for a node.
type DisplayAcl struct {
	Owner  string    `json:"owner"`
	Read   []string  `json:"read"`
	Write  []string  `json:"write"`
	Delete []string  `json:"delete"`
	Public PublicAcl `json:"public"`
}

// PublicAcl represents public access flags.
type PublicAcl struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

// ServerInfo represents the root resource response from the Shock server.
type ServerInfo struct {
	AttributeIndexes []string `json:"attribute_indexes"`
	Contact          string   `json:"contact"`
	ID               string   `json:"id"`
	Auth             []string `json:"auth"`
	AnonPerms        AnonPerms `json:"anonymous_permissions"`
	Resources        []string `json:"resources"`
	ServerTime       string   `json:"server_time"`
	Type             string   `json:"type"`
	URL              string   `json:"url"`
	Uptime           string   `json:"uptime"`
	Version          string   `json:"version"`
}

// AnonPerms represents anonymous user permissions.
type AnonPerms struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

// PreAuthResponse represents a pre-authenticated download URL response.
type PreAuthResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
	Format    string `json:"format"`
	Filename  string `json:"filename"`
	Files     int    `json:"files"`
	Size      int64  `json:"size"`
}

// Response envelopes (unexported).

type nodeResponse struct {
	Status int      `json:"status"`
	Data   *Node    `json:"data"`
	Error  []string `json:"error"`
}

type nodesResponse struct {
	Status     int      `json:"status"`
	Data       []Node   `json:"data"`
	Error      []string `json:"error"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int      `json:"total_count"`
}

type aclResponse struct {
	Status int         `json:"status"`
	Data   *DisplayAcl `json:"data"`
	Error  []string    `json:"error"`
}

type locationsResponse struct {
	Status int        `json:"status"`
	Data   []Location `json:"data"`
	Error  []string   `json:"error"`
}

type preAuthResponseEnvelope struct {
	Status int              `json:"status"`
	Data   *PreAuthResponse `json:"data"`
	Error  []string         `json:"error"`
}

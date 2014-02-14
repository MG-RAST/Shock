package node

import (
	"github.com/MG-RAST/golib/go-uuid/uuid/"
	"encoding/json"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"os"
	"path/filepath"
)

// has
func (node *Node) HasFile() bool {
	if node.File.Name == "" && node.File.Size == 0 && len(node.File.Checksum) == 0 && node.File.Path == "" {
		return false
	}
	return true
}

func (node *Node) HasIndex(index string) bool {
	if virtIdx[index] {
		return true
	} else {
		if node.HasFile() {
			if _, err := os.Stat(node.IndexPath() + "/" + index); err == nil {
				return true
			}
		}
	}
	return false
}

func (node *Node) HasParent() bool {
	for _, linkage := range node.Linkages {
		if linkage.Type == "parent" {
			return true
		}
	}
	return false
}

// path
func (node *Node) Path() string {
	return getPath(node.Id)
}

func (node *Node) IndexPath() string {
	return getIndexPath(node.Id)
}

func (node *Node) FilePath() string {
	if node.File.Path != "" {
		return node.File.Path
	}
	return getPath(node.Id) + "/" + node.Id + ".data"
}

// misc
func (node *Node) setId() {
	node.Id = uuid.New()
	return
}

func (node *Node) FileExt() string {
	if node.File.Name != "" {
		return filepath.Ext(node.File.Name)
	}
	return ""
}

func (node *Node) ToJson() (s string, err error) {
	m, err := json.Marshal(node)
	s = string(m)
	return
}

func contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func getPath(id string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", conf.Conf["data-path"], id[0:2], id[2:4], id[4:6], id)
}

func getIndexPath(id string) string {
	return fmt.Sprintf("%s/idx", getPath(id))
}

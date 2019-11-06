package node

import (
	"net/http"
	"strings"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
)

// POST: /node
func (cr *NodeController) Create(ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// public user
	if u == nil {
		if conf.ANON_WRITE {
			u = &user.User{Uuid: "public"}
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Parse uploaded form
	// all POSTed files writen to temp dir
	params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
	//fmt.Println("params:")
	//spew.Dump(params)
	// clean up temp dir !!
	defer file.RemoveAllFormFiles(files)
	if err != nil {
		if !strings.Contains(err.Error(), http.ErrNotMultipart.ErrorString) {
			// Some error other than request encoding. Theoretically
			// could be a lost db connection between user lookup and parsing.
			// Blame the user, Its probaby their fault anyway.
			err_msg := "err@node_Create: unable to parse form: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

		// If not multipart/form-data it will try to read the Body of the
		// request. If the Body is not empty it will create a file from
		// the Body contents. If the Body is empty it will create an empty
		// node.
		if ctx.HttpRequest().ContentLength != 0 {
			params, files, err = request.DataUpload(ctx.HttpRequest())
			if err != nil {
				err_msg := "err@node_Create: (request.DataUpload) " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
		}

		n, cn_err := node.CreateNodeUpload(u, params, files)

		if cn_err != nil {
			err_msg := "err@node_Create: (node.CreateNodeUpload) " + cn_err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
		if n == nil {
			// Not sure how you could get an empty node with no error
			// Assume it's the user's fault
			err_msg := "err@node_Create: could not create node"
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		} else {
			return responder.RespondWithData(ctx, n)
		}

	}

	// special case, create preauth download url from list of ids
	if _, hasDownloadUrl := params["download_url"]; hasDownloadUrl {
		if idStr, hasIds := params["ids"]; hasIds {
			idList := strings.Split(idStr, ",")
			// validate id list
			var nodeIds []string
			var totalBytes int64
			for _, id := range idList {
				// check if node exists
				n, err := node.Load(id)
				if err != nil {
					if err == mgo.ErrNotFound {
						logger.Error("err@node_Create: (download_url) (node.Load) id=" + id + ": " + e.NodeNotFound)
						return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
					} else {
						// In theory the db connection could be lost between
						// checking user and load but seems unlikely.
						err_msg := "err@node_Create (download_url) (node.Load) id=" + id + ": " + err.Error()
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
					}
				}
				// check ACLs
				rights := n.Acl.Check(u.Uuid)
				prights := n.Acl.Check("public")
				if rights["read"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["read"] == false {
					logger.Error("err@node_Create: (download_url) (Authenticate) id=" + id + ": " + e.UnAuth)
					return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
				}
				if n.HasFile() && !n.HasFileLock() {
					nodeIds = append(nodeIds, n.Id)
					totalBytes += n.File.Size
				}
			}
			if len(nodeIds) == 0 {
				err_msg := "err:@node_Create: (download_url) no available files found"
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}
			// add options - set defaults first
			options := map[string]string{}
			options["archive"] = "zip" // default is zip
			if af, ok := params["archive_format"]; ok {
				if archive.IsValidToArchive(af) {
					options["archive"] = af
				}
			}
			preauthId := util.RandString(20)
			if fn, ok := params["file_name"]; ok {
				options["filename"] = fn
			} else {
				options["filename"] = preauthId
			}
			if !strings.HasSuffix(options["filename"], options["archive"]) {
				options["filename"] = options["filename"] + "." + options["archive"]
			}
			// set preauth
			if p, err := preauth.New(preauthId, "download", nodeIds, options); err != nil {
				err_msg := "err:@node_Create: (download_url) " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			} else {
				data := preauth.PreAuthResponse{
					Url:       util.ApiUrl(ctx) + "/preauth/" + p.Id,
					ValidTill: p.ValidTill.Format(time.ANSIC),
					Format:    options["archive"],
					Filename:  options["filename"],
					Files:     len(nodeIds),
					Size:      totalBytes,
				}
				return responder.RespondWithData(ctx, data)
			}
		}
	}
	// special case, creates multiple nodes
	if archiveId, hasArchiveNode := params["unpack_node"]; hasArchiveNode {
		ns, err := node.CreateNodesFromArchive(u, params, files, archiveId)
		if err != nil {
			err_msg := "err@node_Create: (node.CreateNodesFromArchive) " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}
		return responder.RespondWithData(ctx, ns)
	}
	// Create node
	n, err := node.CreateNodeUpload(u, params, files)
	if err != nil {
		err_msg := "err@node_Create: (node.CreateNodeUpload) " + err.Error()
		logger.Error(err_msg)
		return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return responder.RespondWithData(ctx, n)
}

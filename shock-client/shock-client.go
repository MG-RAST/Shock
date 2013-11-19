package main

import (
	"code.google.com/p/gopass"
	"encoding/json"
	"fmt"
	"github.com/MG-RAST/Shock/shock-client/conf"
	"github.com/MG-RAST/Shock/shock-client/lib"
	"io"
	"os"
	"strconv"
	"strings"
)

const Usage = `Usage: shock-client <command> [options...] [args..]
Global Options:
    -conf                       Config file location (default ~/.shock-client.cfg)

Commands:
help                            This help message
create [options...]
    -attributes=<i>             JSON formated attribute file
                                Note: Attributes will replace all current attributes
    Mutualy exclusive options:
    -full=<u>                   Path to file
    -parts=<p>                  Number of parts to be uploaded
    -virtual_file=<s>           Comma seperated list of node ids
    -remote_path=<p>            Remote file path


pcreate [options...]
    -full=<u>                   Path to file
    -threads=<i>                number of threads to use for uploading (default 4)

    Note: parallel uploading for the whole file.

update [options...] <id>
    -part=<p> -file=<f>         The part number to be uploaded and path to file
                                Note: parts must be set
    Note: With the inclusion of part update options are the same as create.

get <id>

download [options...] <id> [<output>]
    -index=<i>                  Name of index (must be used with -parts)
    -parts={p}                  Part(s) from index, may be a range eg. 1-10
    -index_options=<o>          Additional index options. Varies by index type

    Note: if output is not present the download will be written to stdout.

pdownload [options...] <id> [<output>]
    -threads=<i>                number of threads to use for downloading (default 4)

    Note: parallel download for the whole file. if output is not present the download will 
          be written to a file named as the shock node id.

acl <add/rm> <all/read/write/delete> <users> <id>
    Note: users are in the form of comma delimited list of email address or uuids 

chown <user> <id>
    Note: user is email address or uuid 

auth show                       Displays username of currently authenticated user
     set                        Prompts for user authentication and store credentials 
     set-token <token>          Stores credentials from token
     unset                      Deletes stored credentials

`

// print help & die
func helpf(e string) {
	fmt.Fprintln(os.Stderr, Usage)
	if e != "" {
		fmt.Fprintln(os.Stderr, "Error: "+e)
	}
	os.Exit(1)
}

func handle(err error) {
	fmt.Fprint(os.Stderr, err.Error())
	os.Exit(1)
}

func handleString(e string) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

func handleToken(err error) {
	if strings.Contains(err.Error(), "no such file or directory") {
		fmt.Fprintln(os.Stderr, "No stored authentication credentials found.\n")
	} else {
		fmt.Fprintln(os.Stderr, "%s\n", err.Error())
	}
	os.Exit(1)
}

func setToken(fatal bool) {
	t := &lib.Token{}
	if err := t.Load(); err != nil {
		if fatal {
			handleToken(err)
		}
	} else {
		lib.SetTokenAuth(t)
	}
}

func acl(action, perm, users, id string) (err error) {
	n := lib.Node{Id: id}
	switch action {
	case "add":
		return n.AclAdd(perm, users)
	case "rm":
		return n.AclRemove(perm, users)
	case "chown":
		return n.AclChown(users)
	default:
		helpf("")
	}
	return
}

func main() {
	if len(os.Args) == 1 || os.Args[1] == "help" {
		helpf("")
	}

	cmd := os.Args[1]
	args := conf.Initialize(os.Args[2:])

	setToken(false)
	switch cmd {
	case "create", "update":
		n := lib.Node{}
		if cmd == "update" {
			if len(args) != 1 {
				helpf("update requires <id>")
			} else {
				n.Id = args[0]
			}
		}
		opts := lib.Opts{}
		if ne(conf.Flags["attributes"]) {
			opts["attributes"] = (*conf.Flags["attributes"])
		}
		if t, err := fileOptions(conf.Flags); err != nil {
			helpf(err.Error())
		} else {
			if t == "part" {
				if cmd == "create" {
					helpf("part option only usable with update")
				}
				if !ne(conf.Flags["file"]) {
					helpf("part option requires file")
				}
				opts["upload_type"] = t
				opts["part"] = (*conf.Flags["part"])
				opts["file"] = (*conf.Flags["file"])
				if err := n.Update(opts); err != nil {
					handleString(fmt.Sprintf("Error updating %s: %s\n", n.Id, err.Error()))
				} else {
					n.PP()
				}
			} else {
				if t != "" {
					opts["upload_type"] = t
					opts[t] = (*conf.Flags[t])
					if cmd == "create" {
						if err := n.Create(opts); err != nil {
							handleString(fmt.Sprintf("Error creating node: %s\n", err.Error()))
						} else {
							n.PP()
						}
					} else {
						if err := n.Update(opts); err != nil {
							handleString(fmt.Sprintf("Error updating %s: %s\n", n.Id, err.Error()))
						} else {
							n.PP()
						}
					}
				} else {
					if err := n.Create(opts); err != nil {
						handleString(fmt.Sprintf("Error creating node: %s\n", err.Error()))
					} else {
						n.PP()
					}
				}
			}
		}
	case "pcreate":
		pcreate(args)

	case "get":
		if len(args) != 1 {
			helpf("get requires <id>")
		}
		n := lib.Node{Id: args[0]}
		if err := n.Get(); err != nil {
			fmt.Printf("Error retrieving %s: %s\n", n.Id, err.Error())
		} else {
			n.PP()
		}
	case "download":
		if len(args) < 1 {
			helpf("download requires <id>")
		}
		index := conf.Flags["index"]
		parts := conf.Flags["parts"]
		indexOptions := conf.Flags["index_options"]
		opts := lib.Opts{}
		if ne(index) || ne(parts) || ne(indexOptions) {
			if ne(index) && ne(parts) {
				opts["index"] = (*index)
				opts["parts"] = (*parts)
				if ne(indexOptions) {
					opts["index_options"] = (*indexOptions)
				}
			} else {
				helpf("index and parts options must be used together")
			}
		}
		n := lib.Node{Id: args[0]}
		if ih, err := n.Download(opts); err != nil {
			fmt.Printf("Error downloading %s: %s\n", n.Id, err.Error())
		} else {
			if len(args) == 3 {
				if oh, err := os.Create(args[1]); err == nil {
					if s, err := io.Copy(oh, ih); err != nil {
						handleString(fmt.Sprintf("Error writing output: %s\n", err.Error()))
					} else {
						fmt.Printf("Success. Wrote %d bytes\n", s)
					}
				} else {
					handleString(fmt.Sprintf("Error writing output: %s\n", err.Error()))
				}
			} else {
				io.Copy(os.Stdout, ih)
			}
		}
	case "pdownload":
		if len(args) < 1 {
			helpf("pdownload requires <id>")
		}
		n := lib.Node{Id: args[0]}
		if err := n.Get(); err != nil {
			handleString(fmt.Sprintf("Error retrieving %s: %s\n", n.Id, err.Error()))
		}

		totalChunk := int(n.File.Size / conf.CHUNK_SIZE)
		m := n.File.Size % conf.CHUNK_SIZE
		if m != 0 {
			totalChunk += 1
		}
		if totalChunk < 1 {
			totalChunk = 1
		}
		splits := conf.DOWNLOAD_THREADS
		if ne(conf.Flags["threads"]) {
			if th, err := strconv.Atoi(*conf.Flags["threads"]); err == nil {
				splits = th
			}
		}
		if totalChunk < splits {
			splits = totalChunk
		}

		fmt.Printf("downloading using %d threads\n", splits)

		var filename string
		if len(args) == 2 {
			filename = args[1]
		} else {
			filename = n.Id
		}

		oh, err := os.Create(filename)
		if err != nil {
			handleString(fmt.Sprintf("Error creating output file %s: %s\n", filename, err.Error()))
			return
		}
		oh.Close()

		ch := make(chan int, 1)
		split_size := totalChunk / splits
		remainder := totalChunk % splits
		//splitting, if total chunk is 10, each split will have 3,3,2,2 chunks respectively
		for i := 0; i < splits; i++ {
			var start, end int
			if i < remainder {
				start = (split_size+1)*i + 1
				end = start + split_size
			} else {
				start = (split_size+1)*remainder + split_size*(i-remainder) + 1
				end = start + split_size - 1
			}
			part_string := fmt.Sprintf("%d-%d", start, end)
			opts := lib.Opts{}
			opts["index"] = "size"
			opts["parts"] = part_string
			opts["index_options"] = fmt.Sprintf("chunk_size=%d", conf.CHUNK_SIZE)

			start_offset := (int64(start) - 1) * conf.CHUNK_SIZE
			go downloadChunk(n, opts, filename, start_offset, ch)
		}
		for i := 0; i < splits; i++ {
			<-ch
		}
	case "auth":
		if len(args) < 1 {
			helpf("auth requires show/set/set-token/unset")
		}
		switch args[0] {
		case "set":
			var username, password string
			fmt.Printf("Please authenticate to store your credentials.\nusername: ")
			fmt.Scan(&username)
			password, _ = gopass.GetPass("password: ")
			if t, err := lib.OAuthToken(username, password); err == nil {
				if err := t.Store(); err != nil {
					handleString(fmt.Sprintf("Authenticated but failed to store token: %s\n", err.Error()))
				}
				fmt.Printf("Authenticated credentials stored for user %s. Expires in %d days.\n", t.UserName, t.ExpiresInDays())
			} else {
				fmt.Printf("%s\n", err.Error())
			}
		case "set-token":
			if len(args) != 2 {
				helpf("auth set-token requires token.")
			}
			t := &lib.Token{}
			if err := json.Unmarshal([]byte(args[1]), &t); err != nil {
				handleString("Invalid auth token.\n")
			}
			if err := t.Store(); err != nil {
				handleString(fmt.Sprintf("Failed to store token: %s\n", err.Error()))
			}
			fmt.Printf("Authenticated credentials stored for user %s. Expires in %d days.\n", t.UserName, t.ExpiresInDays())
		case "unset":
			t := &lib.Token{}
			if err := t.Delete(); err != nil {
				fmt.Printf("%s\n", err.Error())
			} else {
				fmt.Printf("Stored authentication credentials have been deleted.\n")
			}
		case "show":
			t := &lib.Token{}
			if err := t.Load(); err != nil {
				handleToken(err)
			} else {
				fmt.Printf("Authenticated credentials stored for user %s. Expires in %d days.\n", t.UserName, t.ExpiresInDays())
			}
		}
	case "acl":
		if len(args) != 4 {
			helpf("acl requires 4 arguments")
		}
		if err := acl(args[0], args[1], args[2], args[3]); err != nil {
			handle(err)
		}
	case "chown":
		if len(args) != 2 {
			helpf("chown requires <user> and <id>")
		}
		if err := acl("chown", "", args[0], args[1]); err != nil {
			handle(err)
		}
	default:
		helpf("invalid command")
	}
}

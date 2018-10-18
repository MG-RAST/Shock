package main

import (
	"flag"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	flags       *flag.FlagSet
	archive     string
	attributes  string
	attrQuery   arrayFlags
	bearer      string
	chunk       string
	column      int
	compression string
	copy        string
	debug       bool
	dir         string
	direction   string
	distinct    string
	expiration  string
	filename    string
	filepath    string
	force       bool
	index       string
	length      int
	limit       int
	md5sum      bool
	offset      int
	order       string
	otherQuery  arrayFlags
	output      string
	part        int
	parts       string
	pretty      bool
	remote      string
	seek        int
	shock_url   string
	token       string
	unexpire    bool
	virtual     string
)

const USAGE = `Usage: shock-client <command> [options...] [args..]

Global Options:
    --shock_url=<url>           URL of shock-server, default is http://localhost:7445, or env SHOCK_URL
    --token=<s>                 User token for authorization, or env TOKEN
    --bearer=<s>                Bearer token, default is 'mgrast', or env BEARER
    --output=<p>                File to write ouput too, default is stdout
    --pretty                    JSON output is pretty-print format, default is condensed
    --debug                     run shock client in debug mode

Commands:

help / --help / -h              This help message

info
    Note: returns shock-server info from base url

create [options...]
    --expiration=<s>            Epiration time: use intiger followed by unit [0-9]+[MHD], default is no expiration
    --filename=<s>              Filename to use, default is upload-file basename
    --attributes=<p>            JSON formated attribute file, with --chunk option, applied at end of upload
                                Note: Attributes will replace all current attributes
    
    --dir=<p                    Directory for temp files, used with --chunk
    --chunk=<s>                 Upload size of each chunk, able to resume if disconnected
                                Use integer followed by unit: [0-9]+[KMG]
                                Note: this uses parts node, runs mutliple shock uploads,
                                      only used with --filepath option
    
    --compression=<gzip|bzip2>  Uncompress after upload, default is off
                                Note: used with --filepath and --part options
    
    Mutualy exclusive options:
    --filepath=<p>              Path to file
    --part=<s>                  Number of parts to be uploaded
    --virtual=<s>               Comma seperated list of node ids
    --remote=<url>              URI of remote file
    --copy=<id>                 ID of node to copy

update [options...] <id>
    Note: options are the same as create with additional below
    --unexpire                  Remove expiration
    --part=<s> --filepath=<p>   The part number to be uploaded and path to file
                                Note: parts must be set

resume [options...] <id>
    Note: this is for resuming an incomplete upload using: creatut/update --chunk
    --attributes=<p>            JSON formated attribute file, applied at end up upload
    --filepath=<p>              Path to file
    --dir=<p                    Directory for temp files

unpack [options] <id>
    Note: creates mutliple nodes from a parent archive format node
    --attributes=<p>            JSON formated attribute file
                                Note: same attributes applied to all created nodes
    --archive=<tar|tar.gz|tar.bz2|zip>

index [options...] <id> <line|column|record|chunkrecord|bai>
    --force                     Forces index to be rebuilt if exists
    --column=<i>                Column number to index by if using 'column' type

delete [options...] <id>

get [options...] <id>

query [options...] 
    --limit=<i>                 Max number of output nodes, default is 25
    --offset=<i>                Position to start rerieving nodes at, default is 0
    --order=<s>                 Field to sort nodes by, default is 'created_on'
    --direction=<asc|desc>      Direction to sort by, default is 'desc'
    
    Query field and values, may be used more than once each
    Note: use "FIELD_NAME:VALUE" with no whitespace
    --attribute=<field:value>   Attribute field and value to search with, may be used more than once
    --other=<field:value>       Any non-attributes node field and value to search with, may be used more than once
    
    Note: this option returns a list of fields names and not nodes
    --distinct=<s>              An attribute field name, returns all unique values in that field

download [options...] <id>
    --md5                       Create md5 checksum of downloaded file
                                Note: requires output file
    Mutualy exclusive options:
    --index=<s> --parts={s}      Name of index and part(s) to retrieve, may be a range eg. 1-10
    --seek=<i> --length=<i>     Download bytes from file for given seek and length

acl [options...] <id> <get|add|delete> <all|read|write|delete> <users>
    Note: users are in the form of comma delimited list of user-names or uuids,
          not required for get action

public [options...] <id> <add|delete>

chown [options...] <id> <user>
    Note: user is user-name or uuid 

`

func setFlags() (flags *flag.FlagSet) {
	flags = flag.NewFlagSet("shockclient", flag.ContinueOnError)
	flags.StringVar(&archive, "archive", "", "")
	flags.StringVar(&attributes, "attributes", "", "")
	flags.Var(&attrQuery, "attribute", "")
	flags.StringVar(&bearer, "bearer", "mgrast", "")
	flags.StringVar(&chunk, "chunk", "", "")
	flags.IntVar(&column, "column", 0, "")
	flags.StringVar(&compression, "compression", "", "")
	flags.StringVar(&copy, "copy", "", "")
	flags.BoolVar(&debug, "debug", false, "")
	flags.StringVar(&dir, "dir", ".", "")
	flags.StringVar(&direction, "direction", "", "")
	flags.StringVar(&distinct, "distinct", "", "")
	flags.StringVar(&expiration, "expiration", "", "")
	flags.StringVar(&filename, "filename", "", "")
	flags.StringVar(&filepath, "filepath", "", "")
	flags.BoolVar(&force, "force", false, "")
	flags.StringVar(&index, "index", "", "")
	flags.IntVar(&length, "length", 0, "")
	flags.IntVar(&limit, "limit", 0, "")
	flags.BoolVar(&md5sum, "md5", false, "")
	flags.IntVar(&offset, "offset", 0, "")
	flags.StringVar(&order, "order", "", "")
	flags.Var(&otherQuery, "other", "")
	flags.StringVar(&output, "output", "stdout", "")
	flags.IntVar(&part, "part", 0, "")
	flags.StringVar(&parts, "parts", "", "")
	flags.BoolVar(&pretty, "pretty", false, "")
	flags.StringVar(&remote, "remote", "", "")
	flags.IntVar(&seek, "seek", -1, "")
	flags.StringVar(&shock_url, "shock_url", "http://localhost:7445", "")
	flags.StringVar(&token, "token", "", "")
	flags.BoolVar(&unexpire, "unexpire", false, "")
	flags.StringVar(&virtual, "virtual", "", "")
	return
}

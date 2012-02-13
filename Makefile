include ~/go/src/Make.inc

TARG=bin/shock-server

GOFILES=\
        shock-server.go\
		lib/Node.go\
		lib/NodeRoutes.go\

include ~/go/src/Make.cmd
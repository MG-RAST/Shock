include $(GOROOT)/src/Make.inc

TARG=bin/shock-server

GOFILES=\
        shock-server.go\
		lib/node.go\
		lib/nodeRoutes.go\

include $(GOROOT)/src/Make.cmd

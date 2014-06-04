# Makefile for Shock
#
# Author: Jared Bischof
# 	site: https://github.com/MG-RAST/Shock
# 	
# Targets:
# 	install: Installs the code to the GOPATH directory, sets the version number in main.go
# 	         and reruns a build so the version number appears properly in the compiled code.
#	all: Runs 'go fmt' to format the code, sets the version number in main.go and reruns a
#	     build so the version number appears properly in the compiled code. If this target
#	     is run without having run 'make install' or 'make get', info msg is printed.
#	build: Runs 'go install'
#	fmt: Runs 'go fmt'
#	get: Runs 'go get'
#	clean: Removes src/, pkg/ and bin/ directories inside of GOPATH directory

.PHONY: build

SRCDIR := github.com/MG-RAST/Shock

ifneq ("$(wildcard $(GOPATH)/src/$(SRCDIR))","")
ALL_TARGETS = fmt build
else
ALL_TARGETS = print_info
endif

all: $(ALL_TARGETS)
install: get build

print_info:
	@echo "Please run 'make install' first to retrieve and build the code. 'make all' only rebuilds the binaries once you have the code downloaded."

build: version docs
	go install -v $(SRCDIR)/...

fmt:
	go fmt $(SRCDIR)/...

get:
	go get -v $(SRCDIR)/...
	git clone https://$(SRCDIR).wiki $(GOPATH)/src/$(SRCDIR)/shock-server/site/wiki
	mv $(GOPATH)/src/$(SRCDIR)/shock-server/site/wiki/Home.md $(GOPATH)/src/$(SRCDIR)/shock-server/site/index.md
	cp $(GOPATH)/src/$(SRCDIR)/shock-server/site/wiki/* $(GOPATH)/src/$(SRCDIR)/shock-server/site/

version:
	for i in `cat src/$(SRCDIR)/VERSION`; do sed "s/\[% VERSION %\]/$$i/" < src/$(SRCDIR)/shock-server/main.go > src/$(SRCDIR)/shock-server/main.version.go ; done
	mv src/$(SRCDIR)/shock-server/main.version.go src/$(SRCDIR)/shock-server/main.go

docs:
	@echo '#Shock wiki\n\n[Home](index.md)' > $(GOPATH)/src/$(SRCDIR)/shock-server/site/navigation.md
	for i in `ls $(GOPATH)/src/$(SRCDIR)/shock-server/site/wiki`; do echo "[$$i]($$i)" | sed "s/\.md//" >> $(GOPATH)/src/$(SRCDIR)/shock-server/site/navigation.md; done

clean:
	rm -rf $(GOPATH)/src/github.com/MG-RAST/Shock $(GOPATH)/bin/shock-server $(GOPATH)/bin/shock-client

SUBDIRS := $(wildcard examples/*)


.PHONY: dep
dep:
	go get -u github.com/golang/dep/cmd/dep && \
	dep ensure

.PHONY: build
build:
	go build

.PHONY : examples $(SUBDIRS)
examples : $(SUBDIRS)

$(SUBDIRS) :
	cd $@ && \
	go build

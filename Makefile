include $(GOROOT)/src/Make.inc

TARG=github.com/griffy/birdbrain
GOFMT=gofmt -s -spaces=true -tabindent=false -tabwidth=4

GOFILES=\
  birdbrain.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w ${GOFILES}


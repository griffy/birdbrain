include $(GOROOT)/src/Make.inc

TARG=github.com/griffy/birdbrain/store
GOFMT=gofmt -s -spaces=true -tabindent=false -tabwidth=4

GOFILES=\
  redis.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w ${GOFILES}


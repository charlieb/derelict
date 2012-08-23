include $(GOROOT)/src/Make.inc

TARG=test
GOFILES=main.go

GC=${O}g $(INCLUDES)
LD=${O}l $(LIBS)

include $(GOROOT)/src/Make.cmd

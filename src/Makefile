RACE ?= -race
RUN ?=

.PHONY: all kvsrv1-build kvsrv1 lock1

all: kvsrv1 lock1

kvsrv1-build:
	go build $(RACE) -o main/kvsrv1d main/kvsrv1d.go

kvsrv1: kvsrv1-build
	cd kvsrv1 && go test -v $(RACE) $(RUN)

lock1: kvsrv1-build
	cd kvsrv1/lock && go test -v $(RACE) $(RUN)

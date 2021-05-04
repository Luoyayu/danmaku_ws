GO = go
OUT = build/danmaku
SRC = cmd/main.go

.PHONY: build 
build: 
	${GO} build -o ${OUT} -ldflags "-w -s -buildid=" ${SRC}
	@chmod +x ${OUT}

install: build
	cp ${OUT} ${GOPATH}/bin

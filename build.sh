#!/bin/bash

export CGO_ENABLED=1

go build -ldflags '-linkmode external -extldflags "-static"' -tags sqlite_omit_load_extension -o bin/hub -trimpath -buildvcs=false .
go build -ldflags '-linkmode external -extldflags "-static"' -tags sqlite_omit_load_extension -o bin/agent -trimpath -buildvcs=false ./agent

chown -R 1000:1000 ./bin
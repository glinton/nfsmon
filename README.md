# nfsmon
[![Build Status](https://travis-ci.org/glinton/nfsmon.svg?branch=master)](https://travis-ci.org/glinton/nfsmon)
[![GoDoc](https://godoc.org/github.com/glinton/nfsmon?status.svg)](https://godoc.org/github.com/glinton/nfsmon)

nfsmon provides simple monitoring for NFS mounts. When a stale mount is detected, a user defined remount function is called.

#### Example Use:
```go
package main

import "context"
import "github.com/glinton/nfsmon"

func main() {
  ctx := context.Background()
  nfsmon.Watch(ctx)
}
```

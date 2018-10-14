# lzss

[![GoDoc](https://godoc.org/github.com/blacktop/lzss?status.svg)](https://godoc.org/github.com/blacktop/lzss) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

> LZSS compression package for Go.

---

## Install

```bash
go get github.com/blacktop/lzss
```

## Examples

```golang
import (
    "io/ioutil"

    "github.com/blacktop/lzss"
    "github.com/pkg/errors"
)

func main() {
    c, err := Open("compressed.bin")
    if err != nil {
        return errors.Wrap(err, "failed to open compressed file")
    }
    decompressed := lzss.Decompress(c.Read())
    err = ioutil.WriteFile("compressed.bin.decompressed", decompressed[:UncompressedSize], 0644)
    if err != nil {
        return errors.Wrap(err, "failed to decompress file")
    }
}
```

> **NOTE:** I believe lzss expects the data to be word aligned.

## Credit

Converted to Golang from `BootX-81//bootx.tproj/sl.subproj/lzss.c`

## TODO

- [ ] add Compress func

## License

MIT Copyright (c) 2018 blacktop

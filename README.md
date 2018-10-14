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
    "log"

    "github.com/blacktop/lzss"
    "github.com/pkg/errors"
)

func main() {
    dat, err := ioutil.ReadFile("compressed.bin")
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to read compressed file"))
    }

    decompressed := lzss.Decompress(dat)
    err = ioutil.WriteFile("compressed.bin.decompressed", decompressed, 0644)
    if err != nil {
        log.Fatal(errors.Wrap(err, "failed to decompress file"))
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

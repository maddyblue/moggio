ID3 Parsing For Go
==================

Andrew Scherkus
May 21, 2012


Introduction
------------

Simple ID3 parsing library for go based on the specs at www.id3.org.

It doesn't handle everything but at least gets the imporant bits like artist,
album, track, etc...


Usage
-----
Pass in a suitable io.Reader and away you go!

    fd, _ := os.Open("foo.mp3")
    defer fd.Close()
    file := id3.Read(fd)
    if file != nil {
            fmt.Println(file)
    }


Examples
--------
An example tag reading program can be found under id3/tagreader.

    go install id3/tagreader
    $GOPATH/bin/tagreader path/to/file.mp3 [...]

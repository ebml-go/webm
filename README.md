ebml-go // webm
===============

Package `webm` implements parser, reader and seeker for files in WebM container. 

WebM files consist of video streams compressed with the VP8 or VP9 video codec,
audio streams compressed with the Vorbis or Opus audio codecs. The WebM file structure is based on the Matroska media container.
See [WebM FAQ](http://www.webmproject.org/about/faq/).

The parser uses an [EBML decoder](https://github.com/ebml-go/ebml) for Go programming language.

### Installation

```
$ go get github.com/ebml-go/webm
```

### License

The BSD 3-Clause License.

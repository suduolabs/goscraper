package utils

import (
	"time"

	"github.com/valyala/fasthttp"
)

// Client is the shared HTTP client for every Academia request.
//
// fasthttp's package level Do uses a default client whose ReadBufferSize is
// 4096, and that is not enough for Academia. A session expiry response carries
// a Set-Cookie line per cleared cookie, which pushes the header block past 4 KB
// and makes fasthttp fail with
//
//	small read buffer. Increase ReadBufferSize. Buffer size=4096
//
// before it has parsed a status code. The recoverable case, a session that
// needs renewing, then surfaces as an opaque transport error on every endpoint
// at once, and callers cannot tell it apart from the portal being unreachable.
var Client = &fasthttp.Client{
	ReadBufferSize:      32 * 1024,
	MaxResponseBodySize: 16 * 1024 * 1024,
	ReadTimeout:         30 * time.Second,
	WriteTimeout:        30 * time.Second,
}

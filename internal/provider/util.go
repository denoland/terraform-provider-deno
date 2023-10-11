package provider

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

// encodePath applies URL encoding to the given path, with directory separator
// `/` preserved.
func encodePath(path string) string {
	arr := []string{}
	for _, part := range strings.Split(path, "/") {
		escaped := url.QueryEscape(part)
		arr = append(arr, escaped)
	}
	return strings.Join(arr, "/")
}

func calculateGitSha1(b []byte) string {
	prefix := []byte(fmt.Sprintf("blob %d\x00", len(b)))
	h := sha1.New()
	h.Write(prefix)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

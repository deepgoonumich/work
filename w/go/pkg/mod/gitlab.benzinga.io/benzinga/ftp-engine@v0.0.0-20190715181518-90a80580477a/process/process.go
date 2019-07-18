package process

import (
	"bytes"
	"encoding/hex"
	"log"
	"regexp"

	"github.com/minio/sha256-simd"

	"gitlab.benzinga.io/benzinga/content-models/models"
)

type ContentType string

const (
	Story        ContentType = "story"
	PressRelease ContentType = "press-release"
)

func (c ContentType) String() string {
	return string(c)
}

var ContentTypeMappings = map[string]ContentType{
	"abnewswire":             PressRelease,
	"accesswire_pr":          PressRelease,
	"acnnewswire_story":      PressRelease,
	"businesswire_story":     PressRelease,
	"bz_pr_thomson_reuters":  PressRelease,
	"comtex_story":           PressRelease,
	"globenewswire_story":    PressRelease,
	"marketwire_story":       PressRelease,
	"newswire_pressreleases": PressRelease,
	"newswire_story":         PressRelease,
	"pr_story":               PressRelease,
	"prweb_story":            PressRelease,
	"story":                  Story,
	"webwire_story":          PressRelease,
	"wired_release":          PressRelease,
}

var bodyTickerPath = regexp.MustCompile(`"{1}(/stock{1}/{1}\S*)"{1}`)

// RewriteBodyTickerPaths returns body with Tickers rewritten from relative to absolute URLs
// using the given prefix. Prefix should not have trailing slash.
func RewriteBodyTickerPaths(urlPrefix, body string) string {
	return bodyTickerPath.ReplaceAllString(body, `"`+urlPrefix+"${1}"+`"`)
}

func (o *Output) CalculateChecksumSize() *Output {
	h := sha256.New()
	_, err := h.Write(o.Data.Bytes())
	if err != nil {
		log.Panicln(err)
	}
	o.Checksum = hex.EncodeToString(h.Sum(nil))
	o.Size = o.Data.Len()
	return o
}

type Output struct {
	Filename string
	Checksum string // SHA256 Hex Output
	Data     *bytes.Buffer
	Size     int
}

type Processor interface {
	Convert(*models.Event) (*Output, error)
}

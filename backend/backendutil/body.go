package backendutil

import (
	"bytes"
	"errors"
	"io"

	"github.com/antihax/go-imap"
	"github.com/emersion/go-message"
)

var errNoSuchPart = errors.New("backendutil: no such message body part")

// FetchBodySection extracts a body section from a message.
func FetchBodySection(e *message.Entity, section *imap.BodySectionName) (imap.Literal, error) {
	// First, find the requested part using the provided path
	for i := len(section.Path) - 1; i >= 0; i-- {
		n := section.Path[i]

		mr := e.MultipartReader()
		if mr == nil {
			return nil, errNoSuchPart
		}

		for j := 1; j <= n; j++ {
			p, err := mr.NextPart()
			if err == io.EOF {
				return nil, errNoSuchPart
			} else if err != nil {
				return nil, err
			}

			if j == n {
				e = p
				break
			}
		}
	}

	// Then, write the requested data to a buffer
	b := new(bytes.Buffer)

	// Write the header
	mw, err := message.CreateWriter(b, e.Header)
	if err != nil {
		return nil, err
	}
	defer mw.Close()

	// If the header hasn't been requested, discard it
	if section.Specifier == imap.TextSpecifier {
		b.Reset()
	}

	// Write the body, if requested
	switch section.Specifier {
	case imap.EntireSpecifier, imap.TextSpecifier:
		if _, err := io.Copy(mw, e.Body); err != nil {
			return nil, err
		}
	}

	var l imap.Literal = b
	if section.Partial != nil {
		l = bytes.NewReader(section.ExtractPartial(b.Bytes()))
	}
	return l, nil
}

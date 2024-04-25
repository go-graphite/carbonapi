package gosnowth

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// scanToken values represent individual tokens found by query scanners.
type scanToken uint

// scanToken values used by the scanners to identify tokens.
const (
	tokenEOF scanToken = iota
	tokenIllegal
	tokenOB  // Open bracket
	tokenCB  // Close bracket
	tokenOCB // Open curly bracket
	tokenCCB // Close curly bracket
	tokenColon
	tokenComma
	tokenTagCat
	tokenTagVal
	tokenMetric
	tokenStreamTag
	tokenMeasurementTag
)

// tagType value represent whether a tag is a stream or measurment tag.
type tagType uint

const (
	tagStreamTag = iota
	tagMeasurementTag
)

// Tag values represent stream or measurment tags.
type Tag struct {
	Category string
	Value    string
}

// MetricName values represent metric names.
type MetricName struct {
	CanonicalName   string
	Name            string
	StreamTags      []Tag
	MeasurementTags []Tag
}

// NewMetricName creates and initializes a new metric name value.
func NewMetricName() *MetricName {
	return &MetricName{}
}

// metricScanner represents a lexical scanner for metric names.
type metricScanner struct {
	r *bufio.Reader
}

// newMetricScanner returns a new metric name scanner value.
func newMetricScanner(r io.Reader) *metricScanner {
	return &metricScanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (ms *metricScanner) read() rune {
	ch, _, err := ms.r.ReadRune()
	if err != nil {
		return rune(0) // EOF
	}

	return ch
}

// unread places the previously read rune back on the reader.
func (ms *metricScanner) unread() error { return ms.r.UnreadRune() }

// Scan returns the next token and literal value.
func (ms *metricScanner) scan() (scanToken, string) {
	ch := ms.read()

	switch ch {
	case rune(0):
		return tokenEOF, ""
	case '[':
		return tokenOB, string(ch)
	case ']':
		return tokenCB, string(ch)
	case '{':
		return tokenOCB, string(ch)
	case '}':
		return tokenCCB, string(ch)
	case ':':
		return tokenColon, string(ch)
	case ',':
		return tokenComma, string(ch)
	}

	return tokenIllegal, string(ch)
}

// scanTagSep attempts to scan a tag separator from the scan buffer.
func (ms *metricScanner) scanTagSep() (scanToken, string, error) {
	var buf bytes.Buffer
	if ch := ms.read(); ch == '|' {
		if _, err := buf.WriteRune(ch); err != nil {
			return tokenIllegal, "", fmt.Errorf(
				"unable to write to tag separator buffer: %w", err)
		}

		for i := 0; i < 2; i++ {
			ch := ms.read()
			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", fmt.Errorf(
					"unable to write to tag separator buffer: %w", err)
			}
		}

		switch buf.String() {
		case "|ST":
			return tokenStreamTag, buf.String(), nil
		case "|MT":
			return tokenMeasurementTag, buf.String(), nil
		default:
			return tokenIllegal, "", nil
		}
	} else if ch == rune(0) {
		return tokenEOF, "", nil
	}

	return tokenIllegal, "", nil
}

// peekTagSep checks for a tag separator next in the scan buffer.
func (ms *metricScanner) peekTagSep() (scanToken, string, error) {
	if ch := ms.read(); ch == '|' {
		if err := ms.unread(); err != nil {
			return tokenIllegal, "", fmt.Errorf(
				"unable to unread to scan buffer: %w", err)
		}

		if b, err := ms.r.Peek(3); err == nil {
			switch string(b) {
			case "|ST":
				return tokenStreamTag, string(b), nil
			case "|MT":
				return tokenMeasurementTag, string(b), nil
			default:
				return tokenIllegal, "", nil
			}
		}
	}

	return tokenIllegal, "", nil
}

// scanMetricName consumes the current rune and all contiguous ident runes.
func (ms *metricScanner) scanMetricName() (scanToken, string, error) {
	var buf bytes.Buffer

	for {
		ch := ms.read()
		if ch == '|' { //nolint:nestif
			if err := ms.unread(); err != nil {
				return tokenIllegal, "", fmt.Errorf(
					"unable to unread to scan buffer: %w", err)
			}

			tok, _, err := ms.peekTagSep()
			if err != nil {
				return tokenIllegal, "", fmt.Errorf(
					"unable to peek tag separator from scan buffer: %w", err)
			}

			if tok != tokenIllegal {
				// we have a valid separator done scanning name
				break
			}

			// otherwise write the character
			ch := ms.read()
			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", fmt.Errorf(
					"unable to write to metric name buffer: %w", err)
			}
		} else if ch == rune(0) { // EOF
			break
		} else if _, err := buf.WriteRune(ch); err != nil {
			return tokenIllegal, "", fmt.Errorf(
				"unable to write to metric name buffer: %w", err)
		}
	}

	return tokenMetric, buf.String(), nil
}

// scanTagName attempts to read a tag name token from the scan buffer.
func (ms *metricScanner) scanTagName() (scanToken, string, string, error) {
	var buf bytes.Buffer

	var can bytes.Buffer

	quoted := false

loop:
	for {
		ch := ms.read()
		switch ch {
		case '"':
			quoted = !quoted

			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name buffer: %w", err)
			}

			if _, err := can.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name canonical buffer: %w", err)
			}
		case '\\':
			if quoted { //nolint:nestif
				ch2 := ms.read()
				if ch2 == '"' || ch2 == '\\' {
					if _, err := buf.WriteRune(ch2); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name buffer: %w", err)
					}

					if _, err := can.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name canonical buffer: %w", err)
					}

					if _, err := can.WriteRune(ch2); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name canonical buffer: %w", err)
					}
				} else {
					if err := ms.unread(); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to unread rune from tag name: %w", err)
					}

					if _, err := buf.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name buffer: %w", err)
					}

					if _, err := can.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name canonical buffer: %w", err)
					}
				}
			} else {
				if _, err := buf.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name buffer: %w", err)
				}

				if _, err := can.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name canonical buffer: %w", err)
				}
			}
		case ':':
			if !quoted {
				if err := ms.unread(); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to unread to scan buffer: %w", err)
				}

				break loop
			} else {
				if _, err := buf.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name buffer: %w", err)
				}

				if _, err := can.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name canonical buffer: %w", err)
				}
			}
		case rune(0): // EOF
			break loop
		default:
			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name buffer: %w", err)
			}

			if _, err := can.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name canonical buffer: %w", err)
			}
		}
	}

	return tokenTagCat, buf.String(), can.String(), nil
}

// scanTagValue attempts to read a tag value token from the scan buffer.
func (ms *metricScanner) scanTagValue(
	tt tagType,
) (scanToken, string, string, error) {
	var buf, can bytes.Buffer

	quoted := false

loop:
	for {
		ch := ms.read()
		switch ch {
		case '"':
			quoted = !quoted

			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name buffer: %w", err)
			}

			if _, err := can.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to canonical tag name buffer: %w", err)
			}
		case '\\':
			if quoted { //nolint:nestif
				ch2 := ms.read()
				if ch2 == '"' || ch2 == '\\' {
					if _, err := buf.WriteRune(ch2); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name buffer: %w", err)
					}

					if _, err := can.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to canonical tag name buffer: %w", err)
					}

					if _, err := can.WriteRune(ch2); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to canonical tag name buffer: %w", err)
					}
				} else {
					if err := ms.unread(); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to unread rune from tag name: %w", err)
					}

					if _, err := buf.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to tag name buffer: %w", err)
					}

					if _, err := can.WriteRune(ch); err != nil {
						return tokenIllegal, "", "", fmt.Errorf(
							"unable to write to canonical tag name buffer: %w", err)
					}
				}
			} else {
				if _, err := buf.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name buffer: %w", err)
				}

				if _, err := can.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to canonical tag name buffer: %w", err)
				}
			}
		case ',', ']', '}':
			if !quoted && (ch == ',' || (ch == ']' && tt == tagStreamTag) ||
				(ch == '}' && tt == tagMeasurementTag)) {
				if err := ms.unread(); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to unread to scan buffer: %w", err)
				}

				break loop
			} else {
				if _, err := buf.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to tag name buffer: %w", err)
				}

				if _, err := can.WriteRune(ch); err != nil {
					return tokenIllegal, "", "", fmt.Errorf(
						"unable to write to canonical tag name buffer: %w", err)
				}
			}
		case rune(0): // EOF
			break loop
		default:
			if _, err := buf.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to tag name buffer: %w", err)
			}

			if _, err := can.WriteRune(ch); err != nil {
				return tokenIllegal, "", "", fmt.Errorf(
					"unable to write to canonical tag name buffer: %w", err)
			}
		}
	}

	return tokenTagVal, buf.String(), can.String(), nil
}

// MetricParser values are used to parse metric names and stream tags.
type MetricParser struct {
	s *metricScanner
}

// NewMetricParser returns a new instance of MetricParser.
func NewMetricParser(r io.Reader) *MetricParser {
	return &MetricParser{s: newMetricScanner(r)}
}

// parseTagSet performs the functionality to parse a tag set.
func (mp *MetricParser) parseTagSet(tt tagType) (string, []Tag, error) {
	canonical := strings.Builder{}
	tags := []Tag{}

	switch tt {
	case tagStreamTag:
		canonical.WriteString("|ST[")
	case tagMeasurementTag:
		canonical.WriteString("|MT{")
	default:
		return "", nil, fmt.Errorf("invalid tag type: %v", tt)
	}

	var tok scanToken

	var lit, can string

	var err error

	if tok, lit = mp.s.scan(); tok != tokenOB && tok != tokenOCB {
		return "", nil, fmt.Errorf(
			"parse failure, expecting '[' or '{', got: %s", lit)
	}

	for {
		tag := Tag{}

		tok, lit, can, err = mp.s.scanTagName()
		if err != nil {
			return "", nil,
				fmt.Errorf("unable to parse tag name: %w", err)
		}

		if tok != tokenTagCat {
			return "", nil, fmt.Errorf("expected tag name, got: %s", lit)
		}

		tag.Category = lit
		if strings.HasPrefix(tag.Category, `b"`) &&
			strings.HasSuffix(tag.Category, `"`) {
			val := strings.Trim(tag.Category[1:], `"`)

			b, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding,
				bytes.NewBufferString(val)))
			if err != nil {
				return "", nil, fmt.Errorf(
					"unable to parse base64 tag category: %w", err)
			}

			tag.Category = string(b)
		}

		canonical.WriteString(can)

		if tok, lit = mp.s.scan(); tok != tokenColon {
			return "", nil,
				fmt.Errorf("parse failure, expecting ':' got: %s", lit)
		}

		canonical.WriteString(":")

		tok, lit, can, err = mp.s.scanTagValue(tt)
		if err != nil {
			return "", nil,
				fmt.Errorf("unable to parse tag value: %w", err)
		}

		if tok != tokenTagVal {
			return "", nil,
				fmt.Errorf("expected tag value, got: %s", lit)
		}

		tag.Value = lit
		if strings.HasPrefix(tag.Value, `b"`) &&
			strings.HasSuffix(tag.Value, `"`) {
			val := strings.Trim(tag.Value[1:], `"`)

			b, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding,
				bytes.NewBufferString(val)))
			if err != nil {
				return "", nil, fmt.Errorf(
					"unable to parse base64 stream tag value: %w", err)
			}

			tag.Value = string(b)
		}

		canonical.WriteString(can)

		tags = append(tags, tag)

		tok, lit = mp.s.scan()
		if tok == tokenComma {
			// there are additional tags
			canonical.WriteString(",")

			continue
		}

		if tt == tagStreamTag && tok == tokenCB {
			// done with stream tags
			canonical.WriteString("]")

			break
		}

		if tt == tagMeasurementTag && tok == tokenCCB {
			// done with measurement tags
			canonical.WriteString("}")

			break
		}

		return "", nil, fmt.Errorf(
			"should have ',' or ']', or '}', got: %s", lit)
	}

	return canonical.String(), tags, nil
}

// Parse scans and parses a metric name value.
func (mp *MetricParser) Parse() (*MetricName, error) {
	canonical := strings.Builder{}
	metricName := NewMetricName()

	tok, lit, err := mp.s.scanMetricName()
	if err != nil {
		return nil, fmt.Errorf("unable to scan metric name token: %w", err)
	}

	if tok != tokenMetric {
		return nil, fmt.Errorf("expected metric identifier, got: %s ", lit)
	}

	canonical.WriteString(lit)
	metricName.Name = lit

	for {
		// Get any tags in the metric name.
		tok, _, err = mp.s.scanTagSep()
		if err != nil {
			return nil, fmt.Errorf("unable to scan metric tag token: %w", err)
		}

		if tok == tokenEOF {
			break
		}

		if tok == tokenStreamTag {
			can, tags, err := mp.parseTagSet(tagStreamTag)
			if err != nil {
				return nil, err
			}

			canonical.WriteString(can)

			metricName.StreamTags = append(metricName.StreamTags, tags...)
		} else if tok == tokenMeasurementTag {
			can, tags, err := mp.parseTagSet(tagMeasurementTag)
			if err != nil {
				return nil, err
			}

			canonical.WriteString(can)
			metricName.MeasurementTags = append(metricName.MeasurementTags,
				tags...)
		}
	}

	metricName.CanonicalName = canonical.String()

	return metricName, nil
}

// ParseMetricName takes a canonical metric name as a string and parses it
// into a MetricName value containing separated stream tags.
func ParseMetricName(name string) (*MetricName, error) {
	p := NewMetricParser(bytes.NewBufferString(name))

	return p.Parse()
}

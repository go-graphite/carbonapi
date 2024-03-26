package main

import (
	"bytes"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (cfg *listener) tagsAutocompleteHandler(wr http.ResponseWriter, req *http.Request, isValues bool) {
	_ = req.ParseMultipartForm(16 * 1024 * 1024)
	hdrs := make(map[string][]string)

	for n, v := range req.Header {
		hdrs[n] = v
	}

	logger := cfg.logger.With(
		zap.String("function", "findHandler"),
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.Any("form", req.Form),
		zap.Any("headers", hdrs),
	)
	logger.Info("got request")

	if cfg.Code != http.StatusOK {
		wr.WriteHeader(cfg.Code)
		return
	}

	format, err := getFormat(req)
	if err != nil {
		wr.WriteHeader(http.StatusBadRequest)
		_, _ = wr.Write([]byte(err.Error()))
		return
	}

	url := req.URL.String()

	logger.Info("request details",
		zap.Any("query", req.Form),
	)

	returnCode := http.StatusOK
	var tags []string
	if response, ok := cfg.Expressions[url]; ok {
		if response.ReplyDelayMS > 0 {
			delay := time.Duration(response.ReplyDelayMS) * time.Millisecond
			time.Sleep(delay)
		}
		if response.Code == http.StatusNotFound {
			returnCode = http.StatusNotFound
		} else if response.Code != 0 && response.Code != http.StatusOK {
			returnCode = response.Code
			http.Error(wr, http.StatusText(returnCode), returnCode)
			return
		} else {
			tags = response.Tags
		}
	}

	if returnCode == http.StatusNotFound {
		// return 404 when no data
		http.Error(wr, http.StatusText(returnCode), returnCode)
		return
	}

	logger.Info("will return", zap.Strings("response", tags))

	var b []byte
	switch format {
	case jsonFormat:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, t := range tags {
			if i == 0 {
				buf.WriteByte('"')
			} else {
				buf.WriteString(`, "`)
			}
			buf.WriteString(t)
			buf.WriteByte('"')
		}
		buf.WriteByte(']')
		b = buf.Bytes()
		wr.Header().Set("Content-Type", contentTypeJSON)
	default:
		returCode := http.StatusBadRequest
		http.Error(wr, http.StatusText(returnCode), returCode)
		return
	}

	_, _ = wr.Write(b)
}

func (cfg *listener) tagsValuesHandler(wr http.ResponseWriter, req *http.Request) {
	cfg.tagsAutocompleteHandler(wr, req, true)
}

func (cfg *listener) tagsNamesHandler(wr http.ResponseWriter, req *http.Request) {
	cfg.tagsAutocompleteHandler(wr, req, false)
}

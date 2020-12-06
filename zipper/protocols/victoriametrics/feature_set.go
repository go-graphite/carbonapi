package victoriametrics

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type vmSupportedFeatures struct {
	vmVersion                     string
	versionParsed                 [3]int64
	SupportGraphiteFindAPI        bool
	SupportGraphiteTagsAPI        bool
	GraphiteTagsAPIRequiresDedupe bool
}

// Example: v1.46.0
func versionToFeatureSet(logger *zap.Logger, version string) *vmSupportedFeatures {
	logger = logger.With(zap.String("function", "versionToFeatureSet"))
	res := &vmSupportedFeatures{
		vmVersion: version,
	}

	verSplitted := strings.Split(version, ".")
	if len(verSplitted) < 3 {
		logger.Warn("failed to parse version string",
			zap.Strings("version_array", verSplitted),
			zap.Error(fmt.Errorf("expected at least 3 components")),
		)
		return res
	}
	if verSplitted[0] != "v1" {
		return res
	}
	res.versionParsed[0] = 1

	v2, err := strconv.ParseInt(verSplitted[1], 10, 64)
	if err != nil {
		logger.Warn("failed to parse version string",
			zap.Strings("version_array", verSplitted),
			zap.Error(err),
		)
		return res
	}
	res.versionParsed[1] = v2

	if v2 >= 41 {
		res.SupportGraphiteFindAPI = true
	}

	if v2 >= 47 {
		res.SupportGraphiteTagsAPI = true
		res.GraphiteTagsAPIRequiresDedupe = true
	}

	if v2 >= 50 {
		res.GraphiteTagsAPIRequiresDedupe = false
	}

	return res
}

func parseVMVersion(in []byte, fallbackVersion string) string {
	/*
		Logic:
		Step1:
			Input:
				process_start_time_seconds 1607269249
				process_virtual_memory_bytes 3207864320
				vm_app_version{version="victoria-metrics-20201109-230848-tags-v1.46.0-0-g41813eb8", short_version="v1.46.0"} 1
				vm_allowed_memory_bytes 20197038489
				vm_app_start_timestamp 1607269249

			Output:
				{version="victoria-metrics-20201109-230848-tags-v1.46.0-0-g41813eb8", short_version="v1.46.0"} 1
				vm_allowed_memory_bytes 20197038489
				vm_app_start_timestamp 1607269249

		Step2:
			Input:
				{version="victoria-metrics-20201109-230848-tags-v1.46.0-0-g41813eb8", short_version="v1.46.0"} 1
				vm_allowed_memory_bytes 20197038489
				vm_app_start_timestamp 1607269249

			Output:
				v1.46.0"} 1
				vm_allowed_memory_bytes 20197038489
				vm_app_start_timestamp 1607269249
		Step3:
			Input:
				v1.46.0"} 1
				vm_allowed_memory_bytes 20197038489
				vm_app_start_timestamp 1607269249

			Output:
				v1.46.0
	*/
	tokens := [][]byte{
		[]byte("vm_app_version"),
		[]byte("short_version=\""),
		[]byte("\""),
	}

	idx := 0
	for i := range tokens {
		l := 0
		if i != 0 {
			l = len(tokens[i-1])
		}
		in = in[idx+l:]
		idx = bytes.Index(in, tokens[i])
		if idx == -1 {
			return fallbackVersion
		}
	}

	in = in[:idx]
	if len(in) == 0 {
		return fallbackVersion
	}

	return string(in)
}

func (c *VictoriaMetricsGroup) updateFeatureSet(ctx context.Context) {
	logger := c.logger.With(zap.String("function", "updateFeatureSet"))
	rewrite, _ := url.Parse("http://127.0.0.1/metrics")
	var minFeatureSet *vmSupportedFeatures

	res, queryErr := c.httpQuery.DoQueryToAll(ctx, logger, rewrite.RequestURI(), nil)
	if queryErr != nil {
		logger.Debug("got some errors while getting capabilities",
			zap.Error(queryErr),
		)
	}
	if len(res) == 0 {
		return
	}

	for i := range res {
		if res[i] == nil || res[i].Response == nil {
			continue
		}
		version := parseVMVersion(res[i].Response, c.fallbackVersion)
		featureSet := versionToFeatureSet(logger, version)
		if minFeatureSet == nil {
			minFeatureSet = featureSet
			continue
		}

		if minFeatureSet.versionParsed[0] > featureSet.versionParsed[0] ||
			minFeatureSet.versionParsed[1] > featureSet.versionParsed[1] ||
			minFeatureSet.versionParsed[2] > featureSet.versionParsed[2] {
			minFeatureSet = featureSet
		}
	}

	logger.Debug("got feature set",
		zap.Any("featureset", minFeatureSet),
	)

	if minFeatureSet == nil {
		minFeatureSet = versionToFeatureSet(logger, "c.fallbackVersion")
	}

	c.featureSet.Store(minFeatureSet)
}

func (c *VictoriaMetricsGroup) probeVMVersion(ctx context.Context) {
	ticker := time.NewTicker(c.probeVersionInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.updateFeatureSet(ctx)
		}
	}
}

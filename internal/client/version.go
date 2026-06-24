package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

const latestNetbirdReleaseURL = "https://api.github.com/repos/netbirdio/netbird/releases/latest"

var (
	runNetbirdVersion = func(ctx context.Context) ([]byte, error) {
		return exec.CommandContext(ctx, "netbird", "version").Output()
	}
	httpClient = http.DefaultClient
)

type VersionInfo struct {
	CLIVersion      string
	LatestVersion   string
	UpdateAvailable bool
	CheckError      string
}

func NetBirdVersionInfo(ctx context.Context) VersionInfo {
	info := VersionInfo{}

	out, err := runNetbirdVersion(ctx)
	if err != nil {
		info.CheckError = err.Error()
		return info
	}
	info.CLIVersion = strings.TrimSpace(string(out))

	latest, err := latestNetbirdVersion(ctx)
	if err != nil {
		info.CheckError = err.Error()
		return info
	}
	info.LatestVersion = latest
	info.UpdateAvailable = compareVersions(info.CLIVersion, latest) < 0

	return info
}

func latestNetbirdVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestNetbirdReleaseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "netbird-tui")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("latest release check failed: %s", resp.Status)
	}

	var body struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.TagName != "" {
		return strings.TrimPrefix(strings.TrimSpace(body.TagName), "v"), nil
	}
	return strings.TrimPrefix(strings.TrimSpace(body.Name), "v"), nil
}

func compareVersions(a, b string) int {
	aa := versionParts(a)
	bb := versionParts(b)
	max := len(aa)
	if len(bb) > max {
		max = len(bb)
	}
	for i := 0; i < max; i++ {
		av, bv := 0, 0
		if i < len(aa) {
			av = aa[i]
		}
		if i < len(bb) {
			bv = bb[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func versionParts(version string) []int {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	fields := strings.FieldsFunc(version, func(r rune) bool {
		return r == '.' || r == '-' || r == '+'
	})

	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		numeric := strings.Builder{}
		for _, r := range field {
			if r < '0' || r > '9' {
				break
			}
			numeric.WriteRune(r)
		}
		if numeric.Len() == 0 {
			break
		}
		n, err := strconv.Atoi(numeric.String())
		if err != nil {
			break
		}
		parts = append(parts, n)
	}
	return parts
}

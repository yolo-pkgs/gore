package goproxy

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
)

const goProxyURL = "https://proxy.golang.org"

func GetLatestVersion(moduleName string) (string, error) {
	url := fmt.Sprintf("%s/%s/@v/list", goProxyURL, moduleName)
	client := resty.New()

	resp, err := client.R().
		Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("http: %d", resp.StatusCode())
	}

	lines := strings.Fields(string(resp.Body()))
	if len(lines) == 0 {
		return "", errors.New("")
	}

	versions := make([]*version.Version, 0)
	for _, tag := range lines {
		v, err := version.NewVersion(tag)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	sort.Sort(version.Collection(versions))
	lastVersion := versions[len(versions)-1]

	return lastVersion.String(), nil
}

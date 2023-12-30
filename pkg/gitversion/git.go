package gitversion

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
)

func IsGitVersion(ver string) bool {
	if strings.Contains(ver, "devel") {
		return true
	}
	v, err := version.NewVersion(ver)
	if err != nil {
		return false
	}
	segments := v.Segments64()
	for _, segment := range segments {
		if segment != 0 {
			return false
		}
	}
	return true
}

func CloneAndRetrieveLastCommitInfo(repoURL string) (string, time.Time, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git_clone_temp")
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory when done

	// Clone the Git repository to the temporary directory
	cmd := exec.Command("git", "clone", "--no-checkout", "--depth=1", repoURL, tempDir)
	err = cmd.Run()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error cloning repository: %v", err)
	}

	// Get the last commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = tempDir
	output, err := cmd.Output()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error getting last commit hash: %v", err)
	}
	commitHash := string(output)

	// Get the commit time
	cmd = exec.Command("git", "log", "-1", "--format=%ct")
	cmd.Dir = tempDir
	output, err = cmd.Output()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error getting commit time: %v", err)
	}

	// Parse the commit time as a Unix timestamp
	i, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error parsing unix time string to int: %w", err)
	}
	tm := time.Unix(i, 0)

	return commitHash, tm, nil
}

// golang.org/x/pkgsite -> https://go.googlesource.com/pkgsite
func FollowRedirect(client *resty.Client, url string) (string, error) {
	resp, err := client.R().Get(url)
	if err != nil {
		return "", err
	}

	body := string(resp.Body())
	lines := strings.Split(body, "\n")

	// TODO: move away.
	r := regexp.MustCompile(`content.+ git (.*)"`)
	gitURL := ""

	for _, line := range lines {
		if strings.Contains(line, " git ") && strings.Contains(line, "content") {
			parts := r.FindStringSubmatch(line)
			if len(parts) < 2 {
				continue
			}

			gitURL = parts[1]
			break
		}
	}

	if gitURL == "" {
		return "", errors.New("git url not found")
	}

	return gitURL, nil
}

package in

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage"
	"github.com/ljfranklin/test-runner-resource/viewer"
)

var (
	keyRegex = regexp.MustCompile(`/?test-results-(.+)\.xml$`)
	// e.g. "2006-01-02T15:04:05Z"
	timeFormat = time.RFC3339
)

type Getter struct {
	Storage     storage.Storage
	JunitViewer viewer.Junit
}

func (g Getter) Get(request models.InRequest) (models.InResponse, error) {
	startingTimestamp, err := keyToTimestamp(request.Version.Key)
	if err != nil {
		return models.InResponse{}, err
	}

	results, err := g.Storage.List()
	if err != nil {
		return models.InResponse{}, err
	}
	sort.Sort(sort.Reverse(byFileNameRegex(results)))

	keysToFetch := []string{}
	for _, result := range results {
		resultTimestamp, err := keyToTimestamp(result)
		if err != nil {
			return models.InResponse{}, err
		}
		if !resultTimestamp.After(startingTimestamp) {
			keysToFetch = append(keysToFetch, result)
		}
	}

	highestLimit := 0
	for _, summary := range request.Params.Summaries {
		if summary.Limit > highestLimit {
			highestLimit = summary.Limit
		}
	}
	if highestLimit > 0 && len(keysToFetch) > highestLimit {
		keysToFetch = keysToFetch[:highestLimit]
	}

	// TODO: parallelize
	for _, key := range keysToFetch {
		f, err := os.Create(filepath.Join(request.OutputDir, key))
		if err != nil {
			return models.InResponse{}, err
		}

		if err = g.Storage.Get(key, f); err != nil {
			return models.InResponse{}, err
		}

		if err = f.Close(); err != nil {
			return models.InResponse{}, err
		}
	}

	for _, summary := range request.Params.Summaries {
		if err = g.JunitViewer.PrintSummary(summary); err != nil {
			return models.InResponse{}, err
		}
	}

	return models.InResponse{
		Version: request.Version,
		Metadata: map[string]string{
			"test_suite_count": fmt.Sprintf("%d", len(keysToFetch)),
		},
	}, nil
}

type byFileNameRegex []string

func (a byFileNameRegex) Len() int      { return len(a) }
func (a byFileNameRegex) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byFileNameRegex) Less(i, j int) bool {
	iTime, err := keyToTimestamp(a[i])
	if err != nil {
		panic(err)
	}
	jTime, err := keyToTimestamp(a[j])
	if err != nil {
		panic(err)
	}
	return iTime.Before(jTime)
}

// TODO: extract
func keyToTimestamp(key string) (time.Time, error) {
	iMatches := keyRegex.FindStringSubmatch(key)
	if len(iMatches) == 0 {
		return time.Time{}, fmt.Errorf("invalid filename '%s'", key)
	}
	iTime, err := time.Parse(timeFormat, iMatches[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp '%s'", key)
	}

	return iTime, nil
}

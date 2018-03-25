package check

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/ljfranklin/test-runner-resource/models"
	"github.com/ljfranklin/test-runner-resource/storage"
)

var (
	keyRegex = regexp.MustCompile(`/?test-results-(.+)\.xml$`)
	// e.g. "2006-01-02T15:04:05Z"
	timeFormat = time.RFC3339
)

type Checker struct {
	Storage storage.Storage
}

func (c Checker) Check(startingVersion models.Version) (models.CheckResponse, error) {
	var startingTimestamp time.Time
	if startingVersion != (models.Version{}) {
		var err error
		startingTimestamp, err = keyToTimestamp(startingVersion.Key)
		if err != nil {
			return nil, err
		}
	}

	results, err := c.Storage.List()
	if err != nil {
		return nil, err
	}
	sort.Sort(byFileNameRegex(results))

	output := models.CheckResponse{}
	for _, result := range results {
		resultTimestamp, err := keyToTimestamp(result)
		if err != nil {
			return nil, err
		}
		if !startingTimestamp.After(resultTimestamp) {
			output = append(output, models.Version{
				Key: result,
			})
		}
	}

	return output, nil
}

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

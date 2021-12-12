package rabbitmq

import (
	"fmt"
	"net/http"
	"strconv"
)

func validateQueryParams(req *http.Request) (string, int, error) {
	queryParams := req.URL.Query()
	queueNameValues, ok := queryParams["queue"]
	if !ok {
		return "", 0, fmt.Errorf("expected query param: queue")
	}

	c, ok := queryParams["count"]
	if !ok {
		return "", 0, fmt.Errorf("expected query param: count")
	}
	count, err := strconv.Atoi(c[0])
	if err != nil {
		return "", 0, fmt.Errorf("expected count to be a number: %v", err)
	}
	return queueNameValues[0], count, nil
}

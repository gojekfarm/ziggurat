package rabbitmq

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func Test_validateParams(t *testing.T) {
	_, atoiErr := strconv.Atoi("a")

	type test struct {
		name          string
		path          string
		expectedCount int
		expectedQname string
		expectedErr   string
	}
	cases := []test{
		{
			name:          "no query params are sent",
			expectedCount: 0,
			expectedQname: "",
			expectedErr:   "expected query param: queue",
			path:          "/ds",
		},
		{
			name:          "queue param is sent",
			expectedCount: 0,
			expectedQname: "",
			expectedErr:   "expected query param: count",
			path:          "/ds?queue=foo",
		},
		{
			name:          "both params are sent",
			expectedCount: 10,
			expectedQname: "foo",
			expectedErr:   "",
			path:          "/ds?queue=foo&count=10",
		},
		{
			name:          "when count is not a number",
			expectedCount: 0,
			expectedQname: "",
			expectedErr:   fmt.Sprintf("expected count to be a number: %v", atoiErr),
			path:          "/ds?queue=foo&count=a",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, c.path, nil)
			if err != nil {
				t.Errorf("couldn't create a request: %v", err)
			}
			queue, count, err := validateQueryParams(req)
			if queue != c.expectedQname {
				t.Errorf("expected path %s, got %s", c.expectedQname, queue)
			}
			if count != c.expectedCount {
				t.Errorf("expected path %d, got %d", c.expectedCount, count)
			}
			if err == nil {
				if "" != c.expectedErr {
					t.Errorf("expected [%s] to %s", "", c.expectedErr)
				}
			}
			if err != nil && (err.Error() != c.expectedErr) {
				t.Errorf("expected [%s] got [%s]", err.Error(), c.expectedErr)
			}
		})
	}
}

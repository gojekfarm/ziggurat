package statsd

import (
	"context"
	"testing"
	"time"
)

func Test_Run(t *testing.T) {
	type testCase struct {
		Opts     []func(r *runOpts)
		Name     string
		Expected bool
	}

	cases := []testCase{
		{
			Name: "When publish interval is negative",
			Opts: []func(r *runOpts){func(r *runOpts) {
				r.goPublishInterval = -1
			}},
			Expected: false,
		},
		{
			Opts: []func(r *runOpts){func(r *runOpts) {
				r.goPublishInterval = 0
			}},
			Name:     "When publish interval is 0",
			Expected: false,
		},
		{
			Name: "When publish interval is greater than 0",
			Opts: []func(r *runOpts){func(r *runOpts) {
				r.goPublishInterval = 1 * time.Second
			}},
			Expected: true,
		},
		{
			Name:     "When Publish interval is not passed explicitly it sets it to 10 seconds",
			Opts:     []func(opts *runOpts){},
			Expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {

			var called bool

			publishGoRoutines = func(ctx context.Context, interval time.Duration, s *Client) {
				called = true
			}

			s := NewPublisher()

			_ = s.Run(context.Background(), c.Opts...)
			time.Sleep(200 * time.Millisecond)

			if c.Expected != called {
				t.Errorf("expected called to be %v got %v\n", c.Expected, called)
			}

		})
	}

}

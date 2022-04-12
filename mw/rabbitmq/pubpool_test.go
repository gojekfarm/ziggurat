package rabbitmq

import (
	"context"
	"testing"

	"github.com/makasim/amqpextra/publisher"
)

type fakeLogger func()

func (f fakeLogger) Printf(format string, v ...interface{}) {

}

func Test_PoolGetPut(t *testing.T) {
	type test struct {
		GetOps   int
		PutOps   int
		PoolSize int
		Name     string
	}
	cases := []test{{
		GetOps:   4,
		PutOps:   1,
		PoolSize: 5,
		Name:     "pool should have an expected number of items after get and put ops",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			urls := []string{"amqp://user:bitnami@localhost:5672"}
			d, err := newDialer(context.Background(), urls, fakeLogger(func() {}))
			if err != nil {
				t.Fatalf("error creating dialer:%v\n", err)
			}
			cpool, err := newPubPool(context.Background(), c.PoolSize, d, fakeLogger(func() {}))
			if err != nil {
				t.Errorf("error in pool creation:%v\n", err)
			}
			var lastItem *publisher.Publisher
			for i := 0; i < c.GetOps; i++ {
				lastItem, err = cpool.get(context.Background())
				if err != nil {
					t.Fatalf("could not get a publisher from pool:%v\n", err)
				}
			}
			for i := 0; i < c.PutOps; i++ {
				cpool.put(lastItem)
			}

			expected := c.PutOps
			if len(cpool.pool) != expected {
				t.Errorf("expected reminaing items to be %d,got %d\n", expected, len(cpool.pool))
			}
		})
	}
}

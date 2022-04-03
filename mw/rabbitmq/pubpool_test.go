package rabbitmq

import (
    "context"
    "testing"

    "github.com/makasim/amqpextra/publisher"
)

type fakeLogger func()

func (f fakeLogger) Printf(format string, v ...interface{}) {

}

func Test_CpoolCreation(t *testing.T) {
    type test struct {
        Name           string
        PoolSize       int
        ExpectedResult int
        ExpectedErr    error
    }

    cases := []test{
        {
            Name:           "should fill up the pool without errors",
            PoolSize:       5,
            ExpectedResult: 5,
        }}

    for _, c := range cases {
        t.Run(c.Name, func(t *testing.T) {
            ctx, cfn := context.WithCancel(context.Background())
            defer cfn()
            urls := []string{"amqp://user:bitnami@localhost:5672"}

            d, err := newDialer(ctx, urls, fakeLogger(func() {}))
            if err != nil {
                t.Fatalf("error creating dialer:%v\n", err)
            }
            cpool, err := newPubPool(ctx, c.PoolSize, d, fakeLogger(func() {}))
            if err != c.ExpectedErr {
                t.Errorf("expected error to be:%v,got:%v", c.ExpectedErr, err)
            }
            if len(cpool.pool) < c.ExpectedResult {
                t.Errorf("expected:%v,got:%v", c.ExpectedResult, len(cpool.pool))
            }
        })
    }
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
                lastItem = cpool.get()
            }
            for i := 0; i < c.PutOps; i++ {
                cpool.put(lastItem)
            }

            expected := (c.PoolSize - c.GetOps) + c.PutOps
            if len(cpool.pool) != expected {
                t.Errorf("expected reminaing items to be %d,got %d\n", expected, len(cpool.pool))
            }
        })
    }
}

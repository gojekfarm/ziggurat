package rabbitmq

import (
    "context"

    "github.com/makasim/amqpextra"
    "github.com/makasim/amqpextra/logger"
    "github.com/makasim/amqpextra/publisher"
)

type publisherPool struct {
    pool chan *publisher.Publisher
}

func newPubPool(ctx context.Context, size int, d *amqpextra.Dialer, logger logger.Logger) (*publisherPool, error) {
    cpool := &publisherPool{pool: make(chan *publisher.Publisher, size)}
    for i := 0; i < size; i++ {
        p, err := getPublisher(ctx, d, logger)
        if err != nil {
            return nil, err
        }
        cpool.put(p)
    }
    return cpool, nil
}

func (c *publisherPool) get() *publisher.Publisher {
    return <-c.pool
}

func (c *publisherPool) put(p *publisher.Publisher) {
    c.pool <- p
}

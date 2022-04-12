package rabbitmq

import (
	"context"
	"fmt"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/logger"
	"github.com/makasim/amqpextra/publisher"
)

type publisherPool struct {
	pool chan *publisher.Publisher
	d    *amqpextra.Dialer
	l    logger.Logger
}

func newPubPool(ctx context.Context, size int, d *amqpextra.Dialer, logger logger.Logger) (*publisherPool, error) {
	pubPool := &publisherPool{pool: make(chan *publisher.Publisher, size)}
	for i := 0; i < size; i++ {
		p, err := getPublisher(ctx, d, logger)
		if err != nil {
			return nil, err
		}
		pubPool.put(p)
	}
	return pubPool, nil
}

func (c *publisherPool) get(ctx context.Context) (*publisher.Publisher, error) {
	select {
	case pub := <-c.pool:
		return pub, nil
	default:
		pub, err := getPublisher(ctx, c.d, c.l)
		return pub, fmt.Errorf("could not get a channel from publisher pool:%v\n", err)
	}
}

func (c *publisherPool) put(p *publisher.Publisher) {
	select {
	case c.pool <- p:
	default:
		// drop the channel
	}
}

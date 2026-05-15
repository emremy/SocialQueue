package jobs

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

const JobCreatedChannel = "socialqueue:jobs:new"

type Publisher struct {
	redis *goredis.Client
}

func NewPublisher(redis *goredis.Client) *Publisher {
	return &Publisher{
		redis: redis,
	}
}

func (p *Publisher) PublishJobCreated(ctx context.Context) error {
	return p.redis.Publish(ctx, JobCreatedChannel, "new_job").Err()
}

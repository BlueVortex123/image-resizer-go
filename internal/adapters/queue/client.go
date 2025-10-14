package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	return &QueueClient{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

func (queueclient *QueueClient) EnqueueResizeImageTask(url string, width, height int) error {
	payload, err := NewResizeImageTask(url, width, height)
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeResizeImage, payload)
	_, err = queueclient.client.Enqueue(
		task,
		asynq.MaxRetry(3),
		asynq.Timeout(60*time.Second), // 60 seconds
		asynq.Queue("images"),
	)
	return err
}

func (queueclient *QueueClient) Close() error {
	return queueclient.client.Close()
}

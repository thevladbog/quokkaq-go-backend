package jobs

import (
	"encoding/json"
	"os"

	"quokkaq-go-backend/internal/services"

	"github.com/hibiken/asynq"
)

type JobClient interface {
	EnqueueTtsGenerate(payload services.TtsJobPayload) error
	Close() error
}

type jobClient struct {
	client *asynq.Client
}

func NewJobClient() JobClient {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &jobClient{client: client}
}

func (c *jobClient) EnqueueTtsGenerate(payload services.TtsJobPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeTTSGenerate, data)
	_, err = c.client.Enqueue(task, asynq.Queue("default"), asynq.MaxRetry(3))
	return err
}

func (c *jobClient) Close() error {
	return c.client.Close()
}

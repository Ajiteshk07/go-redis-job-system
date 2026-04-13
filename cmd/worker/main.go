package main

import (
	"encoding/json"
	"fmt"
	"job-system/internal/models"
	"job-system/internal/redis"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// 🔥 Scheduler (moves delayed → main queue)
func moveDelayedJobs() {
	for {
		now := float64(time.Now().Unix())

		jobs, _ := redis.Client.ZRangeByScore(redis.Ctx, "delayed_queue", &goredis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%f", now),
		}).Result()

		for _, job := range jobs {
			redis.Client.ZRem(redis.Ctx, "delayed_queue", job)

			redis.Client.ZAdd(redis.Ctx, "job_queue", goredis.Z{
				Score:  float64(time.Now().Unix()),
				Member: job,
			})
		}

		time.Sleep(2 * time.Second)
	}
}

func main() {
	redis.InitRedis()

	go moveDelayedJobs() // 🔥 important

	for {
	jobs, err := redis.Client.ZPopMin(redis.Ctx, "job_queue", 1).Result()
	if err != nil || len(jobs) == 0 {
		time.Sleep(1 * time.Second)
		continue
	}

	var job models.Job
	json.Unmarshal([]byte(jobs[0].Member.(string)), &job)

	fmt.Println("Processing job:", job.ID)

	// ✅ mark processing
	redis.Client.HSet(redis.Ctx, "job:"+job.ID, "status", "processing")

	time.Sleep(2 * time.Second)

	// 🔁 retry logic
	if job.Retries < 2 {
		job.Retries++

		data, _ := json.Marshal(job)

		// push back
		redis.Client.ZAdd(redis.Ctx, "job_queue", goredis.Z{
			Score:  float64(time.Now().Unix()),
			Member: data,
		})

		// update status
		redis.Client.HSet(redis.Ctx, "job:"+job.ID, map[string]interface{}{
			"status":  "retrying",
			"retries": job.Retries,
		})

		// ✅ metric
		redis.Client.Incr(redis.Ctx, "metrics:retry")

		fmt.Println("Retrying job:", job.ID)

		continue
	}

	// ✅ completed
	redis.Client.HSet(redis.Ctx, "job:"+job.ID, "status", "completed")

	// ✅ metric
	redis.Client.Incr(redis.Ctx, "metrics:processed")

	fmt.Println("Completed job:", job.ID)
	}
}
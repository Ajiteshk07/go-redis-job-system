package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9" //

	"job-system/internal/models"
	"job-system/internal/redis"
)

// 🔥 CREATE JOB (with delay)
func createJob(w http.ResponseWriter, r *http.Request) {
	job := models.Job{
		ID:       uuid.New().String(),
		Task:     "demo",
		Status:   "pending",
		Retries:  0,
		Priority: 1,
	}

	delay := 5 * time.Second

	data, _ := json.Marshal(job)

	score := float64(time.Now().Add(delay).Unix())

	// delayed queue
	redis.Client.ZAdd(redis.Ctx, "delayed_queue", goredis.Z{
		Score:  score,
		Member: data,
	})

	// store status
	redis.Client.HSet(redis.Ctx, "job:"+job.ID, map[string]any{
		"status":  "scheduled",
		"retries": 0,
	})

	fmt.Fprintf(w, "Job scheduled: %s\n", job.ID)
}

// 🔥 GET STATUS
func getJobStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	result, _ := redis.Client.HGetAll(redis.Ctx, "job:"+id).Result()

	response, _ := json.Marshal(result)
	w.Write(response)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("x-api-key")

		if apiKey != "secret123" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
func getMetrics(w http.ResponseWriter, r *http.Request) {
	processed, _ := redis.Client.Get(redis.Ctx, "metrics:processed").Result()
	retry, _ := redis.Client.Get(redis.Ctx, "metrics:retry").Result()
	failed, _ := redis.Client.Get(redis.Ctx, "metrics:failed").Result()

	response := map[string]string{
		"processed": processed,
		"retry":     retry,
		"failed":    failed,
	}

	json.NewEncoder(w).Encode(response)
}
func main() {
	redis.InitRedis()

	http.HandleFunc("/create", authMiddleware(createJob))
	http.HandleFunc("/status", authMiddleware(getJobStatus))
	http.HandleFunc("/metrics", authMiddleware(getMetrics))

	fmt.Println("API running on :8080")
	fmt.Println("🔥 THIS VERSION RUNNING")
	http.ListenAndServe(":8080", nil)
}

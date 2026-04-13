# Go Redis Job Queue System

## 🚀 Features
- Background job processing using Redis
- Worker + API architecture
- Dockerized setup
- Retry mechanism
- Real-time job status

## 🛠 Tech Stack
- Golang
- Redis
- Docker

## ▶️ Run Locally
```bash```

docker compose up --build
📌 Endpoints
POST /job → create job
GET /job/:id → check status

---

### ✅ Project Structure 


cmd/
  api/
  worker/
internal/
  handlers/
  models/
  queue/
  redis/


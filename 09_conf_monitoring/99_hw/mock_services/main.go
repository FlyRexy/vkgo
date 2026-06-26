package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// =============================================================================
// Общие утилиты
// =============================================================================

// randDuration возвращает случайную длительность вокруг base с джиттером.
// Среднее = base, но хвосты бывают толстыми.
func randDuration(base time.Duration, jitter float64) time.Duration {
	factor := 1.0 + (rand.Float64()-0.5)*2*jitter
	if factor < 0.1 {
		factor = 0.1
	}
	return time.Duration(float64(base) * factor)
}

// sleepWithSpike — спит примерно base, но иногда (spikeChance%) уходит в spikeDuration.
func sleepWithSpike(base time.Duration, jitter float64, spikeChance float64, spikeDuration time.Duration) {
	if rand.Float64() < spikeChance {
		// Спайк: спим долго
		time.Sleep(spikeDuration + time.Duration(rand.Intn(100))*time.Millisecond)
		return
	}
	time.Sleep(randDuration(base, jitter))
}

// =============================================================================
// Thread Service — порт 15000
// SLA: ~10ms, 100% 200
// Реальность: иногда тормозит (30-200ms), иногда отдаёт 500, иногда таймаутит
// =============================================================================

type threadHandler struct {
	mu      sync.RWMutex
	threads map[string]map[string]interface{} // id -> thread data
}

func newThreadHandler() *threadHandler {
	return &threadHandler{
		threads: make(map[string]map[string]interface{}),
	}
}

func (h *threadHandler) Create(w http.ResponseWriter, r *http.Request) {
	// ---- Имитация нестабильного SLA ----
	// Базовая задержка ~8ms с jitter 50%
	// ~15% запросов получают спайк 50-300ms
	// ~5% запросов получают спайк 300-800ms (2x SLA = 20ms, так что это нарушение)
	// ~3% запросов возвращают 500

	roll := rand.Float64()
	switch {
	case roll < 0.03:
		// 3% — внутренняя ошибка
		time.Sleep(randDuration(5*time.Millisecond, 0.3))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	case roll < 0.08:
		// 5% — огромный спайк 300-800ms (явное нарушение SLA)
		time.Sleep(300*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)
	case roll < 0.23:
		// 15% — средний спайк 30-200ms (нарушение 2x SLA = 20ms)
		time.Sleep(30*time.Millisecond + time.Duration(rand.Intn(170))*time.Millisecond)
	default:
		// 77% — нормальная задержка ~8ms ± jitter
		time.Sleep(randDuration(8*time.Millisecond, 0.5))
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, _ := body["id"].(string)
	if id == "" {
		id = fmt.Sprintf("thread_%d", time.Now().UnixNano())
		body["id"] = id
	}

	h.mu.Lock()
	h.threads[id] = body
	h.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *threadHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Аналогичная логика задержек
	roll := rand.Float64()
	switch {
	case roll < 0.02:
		time.Sleep(randDuration(5*time.Millisecond, 0.3))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	case roll < 0.07:
		time.Sleep(300*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)
	case roll < 0.20:
		time.Sleep(30*time.Millisecond + time.Duration(rand.Intn(170))*time.Millisecond)
	default:
		time.Sleep(randDuration(8*time.Millisecond, 0.5))
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	thread, ok := h.threads[id]
	h.mu.RUnlock()

	if !ok {
		// Тред не найден — но по SLA должны возвращать 200, возвращаем пустой
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": ""})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(thread)
}

// =============================================================================
// Comment Service — порт 16000
// SLA: ~20ms, 100% 200
// Реальность: ещё хуже, чем thread service
// =============================================================================

type commentHandler struct {
	mu       sync.RWMutex
	comments map[string]map[string]interface{} // id -> comment data
	likes    map[string]int                    // commentID -> like count
}

func newCommentHandler() *commentHandler {
	return &commentHandler{
		comments: make(map[string]map[string]interface{}),
		likes:    make(map[string]int),
	}
}

func (h *commentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// ---- Имитация нестабильного SLA ----
	// Базовая задержка ~15ms с jitter 50%
	// ~20% запросов получают спайк 100-500ms
	// ~10% запросов получают спайк 500-1500ms (2x SLA = 40ms, это жёсткое нарушение)
	// ~5% запросов возвращают 500

	roll := rand.Float64()
	switch {
	case roll < 0.05:
		// 5% — внутренняя ошибка
		time.Sleep(randDuration(10*time.Millisecond, 0.3))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	case roll < 0.15:
		// 10% — жёсткий спайк 500-1500ms
		time.Sleep(500*time.Millisecond + time.Duration(rand.Intn(1000))*time.Millisecond)
	case roll < 0.35:
		// 20% — средний спайк 100-500ms
		time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(400))*time.Millisecond)
	default:
		// 65% — нормальная задержка ~15ms ± jitter
		time.Sleep(randDuration(15*time.Millisecond, 0.5))
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	id, _ := body["id"].(string)
	if id == "" {
		id = fmt.Sprintf("comment_%d", time.Now().UnixNano())
		body["id"] = id
	}

	h.mu.Lock()
	h.comments[id] = body
	h.likes[id] = 0
	h.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (h *commentHandler) Like(w http.ResponseWriter, r *http.Request) {
	// Ещё хуже, чем Create
	roll := rand.Float64()
	switch {
	case roll < 0.06:
		// 6% — внутренняя ошибка
		time.Sleep(randDuration(10*time.Millisecond, 0.3))
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	case roll < 0.16:
		// 10% — жёсткий спайк 500-2000ms
		time.Sleep(500*time.Millisecond + time.Duration(rand.Intn(1500))*time.Millisecond)
	case roll < 0.40:
		// 24% — средний спайк 50-500ms
		time.Sleep(50*time.Millisecond + time.Duration(rand.Intn(450))*time.Millisecond)
	default:
		// 60% — нормальная задержка ~15ms ± jitter
		time.Sleep(randDuration(15*time.Millisecond, 0.5))
	}

	cid := r.URL.Query().Get("cid")
	if cid == "" {
		http.Error(w, "missing cid", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	h.likes[cid]++
	h.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// =============================================================================
// Auth (Session) Service — порт 17000
// SLA: ~5ms, 0% 500
// Реальность: иногда отдаёт 500 (!), иногда тормозит
// Этот сервис по SLA должен иметь 0% 500, но реально отдаёт ~5% 500
// =============================================================================

type authHandler struct{}

func newAuthHandler() *authHandler {
	return &authHandler{}
}

func (h *authHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	// ---- Имитация нестабильного SLA ----
	// Базовая задержка ~4ms
	// ~5% запросов возвращают 500 (SLA говорит 0%!)
	// ~8% запросов тормозят 50-200ms (2x SLA = 10ms)

	roll := rand.Float64()
	switch {
	case roll < 0.05:
		// 5% — ошибка 500! (нарушение SLA: должно быть 0%)
		time.Sleep(randDuration(3*time.Millisecond, 0.3))
		http.Error(w, `{"error":"session service unavailable"}`, http.StatusInternalServerError)
		return
	case roll < 0.13:
		// 8% — спайк 50-200ms (2x SLA = 10ms, нарушение)
		time.Sleep(50*time.Millisecond + time.Duration(rand.Intn(150))*time.Millisecond)
	default:
		// 87% — нормальная задержка ~4ms
		time.Sleep(randDuration(4*time.Millisecond, 0.4))
	}

	// Проверяем наличие куки user
	cookie, err := r.Cookie("user")
	if err != nil || cookie.Value == "" {
		// Нет сессии — 403 (это ожидаемое поведение, не нарушение SLA)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Сессия валидна — 200
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// =============================================================================
// Scheduled downtime — имитация падения сервиса в вечернее время (19-21)
// Для опционального задания (3 доп. балла)
// =============================================================================

var (
	downtimeMu     sync.Mutex
	downtimeActive bool
)

func startDowntimeScheduler() {
	go func() {
		for {
			// now := time.Now()
			// hour := now.Hour()
			shouldDown := false

			downtimeMu.Lock()
			downtimeActive = shouldDown
			downtimeMu.Unlock()

			time.Sleep(1 * time.Minute)
		}
	}()
}

func isDowntime() bool {
	downtimeMu.Lock()
	defer downtimeMu.Unlock()
	return downtimeActive
}

// =============================================================================
// Main
// =============================================================================

func main() {
	rand.Seed(time.Now().UnixNano())

	startDowntimeScheduler()

	threadH := newThreadHandler()
	commentH := newCommentHandler()
	authH := newAuthHandler()

	// Thread service — порт 15000
	threadMux := http.NewServeMux()
	threadMux.HandleFunc("/thread", func(w http.ResponseWriter, r *http.Request) {
		if isDowntime() {
			// Во время "падения" — возвращаем 503 с задержкой 1-3s
			time.Sleep(time.Duration(1+rand.Intn(2))*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			threadH.Create(w, r)
		case http.MethodGet:
			threadH.Get(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Comment service — порт 16000
	commentMux := http.NewServeMux()
	commentMux.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		if isDowntime() {
			time.Sleep(time.Duration(1+rand.Intn(2))*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		if r.Method == http.MethodPost {
			commentH.Create(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	commentMux.HandleFunc("/comment/like", func(w http.ResponseWriter, r *http.Request) {
		if isDowntime() {
			time.Sleep(time.Duration(1+rand.Intn(2))*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		if r.Method == http.MethodPost {
			commentH.Like(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Auth service — порт 17000
	authMux := http.NewServeMux()
	authMux.HandleFunc("/int/CheckSession", func(w http.ResponseWriter, r *http.Request) {
		if isDowntime() {
			time.Sleep(time.Duration(1+rand.Intn(2))*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
			http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		if r.Method == http.MethodGet {
			authH.CheckSession(w, r)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	go func() {
		log.Println("Thread service starting on :15000")
		log.Fatal(http.ListenAndServe(":15000", threadMux))
	}()

	go func() {
		log.Println("Comment service starting on :16000")
		log.Fatal(http.ListenAndServe(":16000", commentMux))
	}()

	go func() {
		log.Println("Auth service starting on :17000")
		log.Fatal(http.ListenAndServe(":17000", authMux))
	}()

	log.Println("All mock services started!")
	log.Println("  Thread service :15000 (SLA ~10ms, 100% 200)")
	log.Println("  Comment service:16000 (SLA ~20ms, 100% 200)")
	log.Println("  Auth service   :17000 (SLA ~5ms, 0% 500)")
	log.Println("")
	log.Println("Reality:")
	log.Println("  :15000 — 3% 500, 15% latency spikes 30-200ms, 5% spikes 300-800ms")
	log.Println("  :16000 — 5-6% 500, 20% latency spikes 100-500ms, 10% spikes 500-1500ms")
	log.Println("  :17000 — 5% 500 (!), 8% latency spikes 50-200ms")
	log.Println("")
	log.Println("Downtime mode active between 19:00-21:00 (all services return 503)")

	select {} // block forever
}

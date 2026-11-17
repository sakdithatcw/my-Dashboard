package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	_ "modernc.org/sqlite" // SQLite แบบ pure Go (ไม่ต้องใช้ CGO)
)

var (
	db         *sql.DB
	deviceIDRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
)

// ข้อมูลที่ Pi ส่งเข้ามา (ใช้ pointer เพื่อเช็คได้ว่าส่งฟิลด์มาจริงไหม)
type ReadingIn struct {
	DeviceID    string   `json:"device_id"`
	Temperature *float64 `json:"temperature"`
	Humidity    *float64 `json:"humidity"`
}

type Reading struct {
	DeviceID    string  `json:"device_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
}

type Device struct {
	DeviceID    string  `json:"device_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	LastSeen    string  `json:"last_seen"`
}

const schema = `
CREATE TABLE IF NOT EXISTS readings (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	device_id   TEXT NOT NULL,
	temperature REAL NOT NULL,
	humidity    REAL NOT NULL,
	ts          TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_readings_device ON readings(device_id, id);
`

func main() {
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/readings", handlePostReading)
	mux.HandleFunc("GET /api/readings", handleGetReadings)
	mux.HandleFunc("GET /api/devices", handleGetDevices)

	log.Println("Backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "/data/dashboard.db"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}

	var err error
	db, err = sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1) // SQLite เขียนได้ทีละคน ตั้งให้เรียบง่ายไว้ก่อน

	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}
	log.Println("Database ready:", path)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /api/readings  {"device_id":"pi-01","temperature":28.5,"humidity":65}
func handlePostReading(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var in ReadingIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !deviceIDRe.MatchString(in.DeviceID) {
		http.Error(w, "device_id required (A-Z a-z 0-9 _ . - , max 64 chars)", http.StatusBadRequest)
		return
	}
	if in.Temperature == nil || in.Humidity == nil {
		http.Error(w, "temperature and humidity are required", http.StatusBadRequest)
		return
	}
	t, h := *in.Temperature, *in.Humidity
	if t < -50 || t > 150 {
		http.Error(w, "temperature out of range", http.StatusBadRequest)
		return
	}
	if h < 0 || h > 100 {
		http.Error(w, "humidity must be 0-100", http.StatusBadRequest)
		return
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(
		`INSERT INTO readings (device_id, temperature, humidity, ts) VALUES (?, ?, ?, ?)`,
		in.DeviceID, t, h, ts,
	)
	if err != nil {
		log.Println("insert error:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, Reading{in.DeviceID, t, h, ts})
}

// GET /api/readings?device_id=pi-01&limit=200  (เรียงเก่า -> ใหม่)
func handleGetReadings(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	limit := 200
	if s := r.URL.Query().Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 5000 {
			http.Error(w, "limit must be 1-5000", http.StatusBadRequest)
			return
		}
		limit = n
	}

	rows, err := db.Query(`
		SELECT temperature, humidity, ts FROM (
			SELECT id, temperature, humidity, ts FROM readings
			WHERE device_id = ? ORDER BY id DESC LIMIT ?
		) ORDER BY id ASC`, deviceID, limit)
	if err != nil {
		log.Println("query error:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := make([]Reading, 0)
	for rows.Next() {
		rd := Reading{DeviceID: deviceID}
		if err := rows.Scan(&rd.Temperature, &rd.Humidity, &rd.Timestamp); err != nil {
			log.Println("scan error:", err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out = append(out, rd)
	}
	if err := rows.Err(); err != nil {
		log.Println("rows error:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/devices  -> รายการอุปกรณ์ + ค่าล่าสุดของแต่ละตัว
func handleGetDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
		SELECT r.device_id, r.temperature, r.humidity, r.ts
		FROM readings r
		JOIN (SELECT device_id, MAX(id) AS max_id FROM readings GROUP BY device_id) m
		  ON r.id = m.max_id
		ORDER BY r.device_id`)
	if err != nil {
		log.Println("query error:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := make([]Device, 0)
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.DeviceID, &d.Temperature, &d.Humidity, &d.LastSeen); err != nil {
			log.Println("scan error:", err)
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		log.Println("rows error:", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

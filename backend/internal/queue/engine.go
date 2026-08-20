package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskType mendefinisikan tipe pekerjaan atau tugas asinkron yang diproses worker.
type TaskType string

const (
	TaskTypeExecuteRuleAction    TaskType = "automation.execute_action"
	TaskTypeSendEmailNotification TaskType = "notification.send_email"
	TaskTypeSendWebhookNotification TaskType = "notification.send_webhook"
	TaskTypeTriggerBackup        TaskType = "backup.trigger"
	TaskTypeCleanupTelemetry     TaskType = "telemetry.cleanup"
)

// TaskPayload merepresentasikan struktur pesan pekerjaan asinkron yang dikirim ke antrean.
type TaskPayload struct {
	ID             uuid.UUID       `json:"id"`
	Type           TaskType        `json:"type"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Data           json.RawMessage `json:"data"`
	MaxRetries     int             `json:"max_retries"`
	RetryCount     int             `json:"retry_count"`
	LastError      string          `json:"last_error,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	ExecuteAt      time.Time       `json:"execute_at"`
}

// TaskHandler mendefinisikan fungsi eksekutor untuk memproses satu jenis TaskType tertentu.
// Menerima context dan payload data mentah json.RawMessage.
// Mengembalikan error jika pemrosesan gagal agar dapat dicoba ulang (retry) oleh engine.
type TaskHandler func(ctx context.Context, payload *TaskPayload) error

// QueueEngine mendefinisikan kontrak operasional sistem antrean pekerjaan terdistribusi.
type QueueEngine interface {
	// Enqueue memasukkan tugas baru ke dalam antrean untuk segera diproses oleh worker.
	// Parameter ctx merupakan context pemanggilan.
	// Parameter task merupakan payload pekerjaan yang akan dimasukkan ke antrean.
	// Mengembalikan error jika gagal memasukkan data ke antrean broker.
	Enqueue(ctx context.Context, task *TaskPayload) error

	// EnqueueDelayed memasukkan tugas ke dalam antrean dengan jeda waktu eksekusi tertentu.
	// Parameter ctx merupakan context pemanggilan.
	// Parameter task merupakan payload pekerjaan yang akan dijadwalkan.
	// Parameter delay merupakan durasi waktu tunggu sebelum tugas dipindahkan ke antrean aktif.
	// Mengembalikan error jika gagal menjadwalkan tugas.
	EnqueueDelayed(ctx context.Context, task *TaskPayload, delay time.Duration) error

	// RegisterHandler mendaftarkan fungsi pemroses untuk jenis task tertentu.
	// Parameter taskType merupakan nama jenis tugas.
	// Parameter handler merupakan fungsi pemroses tugas.
	RegisterHandler(taskType TaskType, handler TaskHandler)

	// Start menjalankan consumer pool untuk mulai membaca dan mengeksekusi tugas dari antrean.
	// Parameter ctx merupakan context siklus hidup worker.
	// Mengembalikan error jika inisialisasi antrean gagal.
	Start(ctx context.Context) error

	// Stop menghentikan consumer pool secara anggun (graceful shutdown) setelah tugas yang sedang berjalan selesai.
	// Mengembalikan error jika penutupan koneksi broker bermasalah.
	Stop() error
}

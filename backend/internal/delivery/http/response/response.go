package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type PaginatedMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Data    any           `json:"data"`
	Meta    PaginatedMeta `json:"meta"`
}

// JSON menulis response berformat JSON ke klien HTTP.
// Parameter w merupakan http.ResponseWriter target penulisan output.
// Parameter statusCode menentukan kode status HTTP yang dikirimkan.
// Parameter payload merupakan data yang akan diserialisasi menjadi JSON.
func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// Success menulis response status sukses dengan struktur APIResponse terstandarisasi.
// Parameter w merupakan http.ResponseWriter target penulisan output.
// Parameter statusCode menentukan kode status HTTP sukses (misalnya 200, 201).
// Parameter message berisi pesan deskriptif keberhasilan operasi.
// Parameter data berisi payload data yang dikembalikan ke klien.
func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	JSON(w, statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error menulis response status kegagalan dengan struktur APIResponse terstandarisasi.
// Parameter w merupakan http.ResponseWriter target penulisan output.
// Parameter statusCode menentukan kode status HTTP error (misalnya 400, 404, 500).
// Parameter message berisi pesan deskriptif kesalahan.
// Parameter errDetails berisi rincian atau daftar validasi error.
func Error(w http.ResponseWriter, statusCode int, message string, errDetails any) {
	JSON(w, statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errDetails,
	})
}

// Paginated menulis response status sukses dengan data berpaginasi dan metadata informasi halaman.
// Parameter w merupakan http.ResponseWriter target penulisan output.
// Parameter statusCode menentukan kode status HTTP sukses.
// Parameter message berisi pesan deskriptif keberhasilan operasi.
// Parameter data berisi slice data untuk halaman aktif.
// Parameter page menentukan nomor halaman saat ini (1-indexed).
// Parameter limit menentukan batas jumlah data per halaman.
// Parameter totalItems menentukan total keseluruhan data di database.
func Paginated(w http.ResponseWriter, statusCode int, message string, data any, page, limit int, totalItems int64) {
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	JSON(w, statusCode, PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta: PaginatedMeta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	})
}

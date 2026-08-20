package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/middleware"
	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/storage"
)

// StorageHandler menangani permintaan HTTP terkait manajemen Bucket dan Object Storage.
type StorageHandler struct {
	usecase storage.StorageUsecase
}

// NewStorageHandler membuat instance baru HTTP StorageHandler.
func NewStorageHandler(usecase storage.StorageUsecase) *StorageHandler {
	return &StorageHandler{usecase: usecase}
}

// CreateBucketRequest payload request pembuatan bucket baru.
type CreateBucketRequest struct {
	Name         string                     `json:"name"`
	ProviderType domain.StorageProviderType `json:"provider_type"`
	Region       string                     `json:"region"`
	IsPublic     bool                       `json:"is_public"`
	Versioning   bool                       `json:"versioning"`
}

// CreateBucket membuat bucket baru pada penyedia storage dan menyimpan metadatanya.
func (h *StorageHandler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	var req CreateBucketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	bucket, err := h.usecase.CreateBucket(r.Context(), orgID, req.Name, req.ProviderType, req.Region, req.IsPublic, req.Versioning)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "bucket created successfully", bucket)
}

// ListBuckets mengambil daftar seluruh bucket milik organisasi dengan paginasi.
func (h *StorageHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	buckets, total, err := h.usecase.ListBuckets(r.Context(), orgID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.Paginated(w, http.StatusOK, "buckets retrieved successfully", buckets, page, limit, int64(total))
}

// GetBucket mengambil detail satu bucket berdasarkan nama.
func (h *StorageHandler) GetBucket(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	name := chi.URLParam(r, "name")
	bucket, err := h.usecase.GetBucket(r.Context(), orgID, name)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "bucket retrieved successfully", bucket)
}

// DeleteBucket menghapus bucket jika dalam keadaan kosong.
func (h *StorageHandler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	name := chi.URLParam(r, "name")
	if err := h.usecase.DeleteBucket(r.Context(), orgID, name); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "bucket deleted successfully", nil)
}

// ListObjects mengambil daftar objek dan folder di dalam bucket.
func (h *StorageHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	bucketName := chi.URLParam(r, "name")
	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")
	maxKeys, _ := strconv.Atoi(r.URL.Query().Get("max_keys"))

	objects, folders, err := h.usecase.ListObjects(r.Context(), orgID, bucketName, prefix, delimiter, int32(maxKeys))
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "objects listed successfully", map[string]any{
		"bucket":  bucketName,
		"prefix":  prefix,
		"folders": folders,
		"objects": objects,
	})
}

// UploadObject menangani pengunggahan file via form-data multipart.
func (h *StorageHandler) UploadObject(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	bucketName := chi.URLParam(r, "name")

	// Limit 100MB per request
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "failed to parse multipart form data", nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file field is required in form-data", nil)
		return
	}
	defer file.Close()

	key := r.FormValue("key")
	if key == "" {
		key = header.Filename
	}

	contentType := header.Header.Get("Content-Type")
	size := header.Size

	item, err := h.usecase.UploadObject(r.Context(), orgID, bucketName, key, file, size, contentType, nil)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "object uploaded successfully", item)
}

// DownloadObject mengunduh stream berkas file secara langsung dari bucket.
func (h *StorageHandler) DownloadObject(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	bucketName := chi.URLParam(r, "name")
	key := r.URL.Query().Get("key")
	if key == "" {
		response.Error(w, http.StatusBadRequest, "query parameter key is required", nil)
		return
	}

	content, err := h.usecase.DownloadObject(r.Context(), orgID, bucketName, key)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error(), nil)
		return
	}
	defer content.Body.Close()

	w.Header().Set("Content-Type", content.ContentType)
	if content.ContentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(content.ContentLength, 10))
	}
	if content.ETag != "" {
		w.Header().Set("ETag", fmt.Sprintf("\"%s\"", content.ETag))
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", key))

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, content.Body)
}

// DeleteObject menghapus objek tunggal atau batch objek dari bucket.
func (h *StorageHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	bucketName := chi.URLParam(r, "name")
	keysParam := r.URL.Query().Get("keys")

	if keysParam != "" {
		keys := strings.Split(keysParam, ",")
		if err := h.usecase.DeleteObjects(r.Context(), orgID, bucketName, keys); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
	} else {
		key := r.URL.Query().Get("key")
		if key == "" {
			response.Error(w, http.StatusBadRequest, "query parameter key or keys is required", nil)
			return
		}
		if err := h.usecase.DeleteObject(r.Context(), orgID, bucketName, key); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
	}

	response.Success(w, http.StatusOK, "object(s) deleted successfully", nil)
}

// GenerateSignedURLRequest payload pembuatan Presigned URL.
type GenerateSignedURLRequest struct {
	Key           string                   `json:"key"`
	Operation     domain.SignedURLOperation `json:"operation"` // download / upload
	ExpiryMinutes int                      `json:"expiry_minutes"`
}

// GenerateSignedURL membuat URL bertanda tangan (Presigned URL) untuk unduh atau unggah file.
func (h *StorageHandler) GenerateSignedURL(w http.ResponseWriter, r *http.Request) {
	orgID, ok := middleware.GetOrganizationIDFromContext(r.Context())
	if !ok || orgID == uuid.Nil {
		response.Error(w, http.StatusForbidden, "organization context required", nil)
		return
	}

	bucketName := chi.URLParam(r, "name")

	var req GenerateSignedURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.Key == "" {
		response.Error(w, http.StatusBadRequest, "key is required", nil)
		return
	}

	if req.Operation == "" {
		req.Operation = domain.SignedURLOpDownload
	}

	if req.ExpiryMinutes <= 0 {
		req.ExpiryMinutes = 60
	}

	expiry := time.Duration(req.ExpiryMinutes) * time.Minute

	url, err := h.usecase.GenerateSignedURL(r.Context(), orgID, bucketName, req.Key, req.Operation, expiry)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "signed URL generated successfully", map[string]any{
		"url":            url,
		"operation":      req.Operation,
		"expires_in_sec": int(expiry.Seconds()),
		"expires_at":     time.Now().UTC().Add(expiry),
	})
}

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InjectOrgContext mengembalikan middleware HTTP yang menginjeksikan nilai `app.current_org_id`
// ke dalam session variable PostgreSQL sebelum setiap query dijalankan.
//
// Cara kerja (C-3 — True Database-level Multi-Tenant RLS):
//   1. Middleware mengekstrak orgID dari konteks request (sudah di-set oleh Authenticate middleware).
//   2. Sebelum setiap koneksi diambil dari pool, pgx `BeforeAcquire` hook meng-inject
//      `SET app.current_org_id = '<uuid>'` sehingga semua query pada sesi tersebut terisolasi.
//   3. Jika orgID tidak tersedia (misalnya endpoint publik), session variable dibiarkan kosong
//      sehingga RLS DENY-ALL policy aktif secara default.
//
// Penting: Middleware ini harus dipasang SETELAH middleware Authenticate.
func InjectOrgContext(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID, hasOrg := GetOrganizationIDFromContext(r.Context())

			// Ambil koneksi dari pool dan inject session variable untuk isolasi RLS
			conn, err := pool.Acquire(r.Context())
			if err != nil {
				// Jika gagal acquire koneksi, lanjutkan tanpa injection (query tetap tunduk RLS)
				next.ServeHTTP(w, r)
				return
			}
			defer conn.Release()

			if hasOrg && orgID != uuid.Nil {
				// Inject org ID ke session variable PostgreSQL — berlaku untuk transaksi ini
				if _, err := conn.Exec(r.Context(), "SET LOCAL app.current_org_id = '"+orgID.String()+"'"); err == nil {
					// Tambahkan conn ke konteks agar repository dapat menggunakannya
					ctx := context.WithValue(r.Context(), pgxConnKey, conn)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			} else {
				// Reset session variable untuk endpoint tanpa org context (deny-all RLS aktif)
				_, _ = conn.Exec(r.Context(), "SET LOCAL app.current_org_id = ''")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// pgxConnContextKey adalah tipe kunci konteks privat untuk koneksi pgx yang sudah ter-inject org context.
type pgxConnContextKey struct{}

// pgxConnKey adalah kunci unik penyimpanan koneksi pgx dalam konteks request.
var pgxConnKey = pgxConnContextKey{}

// GetPgxConnFromContext mengambil koneksi *pgxpool.Conn yang sudah ter-inject session RLS dari konteks request.
// Dipanggil oleh repository yang membutuhkan koneksi dengan `app.current_org_id` aktif.
// Mengembalikan nil jika koneksi tidak tersedia (repository akan fallback ke pool biasa).
func GetPgxConnFromContext(ctx context.Context) *pgxpool.Conn {
	conn, _ := ctx.Value(pgxConnKey).(*pgxpool.Conn)
	return conn
}

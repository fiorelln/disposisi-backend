# Disposisi Backend

Backend ini adalah sistem manajemen surat digital untuk SMK 2 Singosari yang menyediakan:

- autentikasi JWT
- manajemen user dan jabatan
- permintaan OTP untuk reset password
- upload surat PDF
- pembuatan dan persetujuan disposisi
- dokumentasi OpenAPI

## Prasyarat

Pastikan sudah terpasang:

- Go 1.20 atau lebih baru
- PostgreSQL
- `git`

## Konfigurasi

Buat file `.env` di root proyek dengan variabel berikut:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=disposisi_db
DB_SSLMODE=disable
JWT_SECRET=supersecretjwt
RESEND_API_KEY=<api-key-resend>
RESEND_EMAIL=<email-pengirim>
```

## Inisialisasi

1. Clone repositori:

   ```bash
   git clone https://example.com/your-repo.git
   cd disposisi-backend
   ```

2. Install dependensi Go:

   ```bash
   go mod tidy
   ```

3. Siapkan database PostgreSQL:

   - Buat database sesuai `DB_NAME`.
   - Pastikan kredensial `.env` cocok.

4. Jalankan aplikasi:

   ```bash
   go run main.go
   ```

Aplikasi akan berjalan di `http://localhost:7000`.

## Endpoint Dokumentasi

- `GET /docs` untuk halaman dokumentasi
- `GET /openapi.yaml` untuk file OpenAPI

## Endpoint Tambahan

- `GET /dashboard` untuk ringkasan dashboard berdasarkan jabatan

## Fitur Utama yang Sudah Selesai

1. Autentikasi dan otorisasi:
   - `POST /auth/login`
   - `POST /admin/users`
   - Login menghasilkan JWT dengan klaim `user_id` dan `roles`

2. Manajemen OTP dan reset password:
   - `POST /auth/forgot-password`
   - `POST /auth/verify-otp`
   - `POST /auth/reset-password`
   - Batas 5 OTP dalam 25 menit dan OTP kadaluarsa 2 menit

3. Upload dan manajemen surat:
   - `POST /surat/upload`
   - `GET /surat/:surat_id`
   - `GET /surat`
   - Validasi file PDF dan ukuran maksimal 5MB

4. Disposisi surat:
   - `POST /disposisi` (hanya role Admin, TU, Kepala TU, Kepala Sekolah)
   - `POST /disposisi/approve` (hanya Kepala Sekolah)
   - `GET /disposisi/surat/:surat_id`
   - `GET /disposisi`

5. Dashboard berdasarkan jabatan:
   - `GET /dashboard`
   - Menampilkan ringkasan surat, disposisi, dan persetujuan yang perlu diproses

6. Kontrol akses berbasis peran:
   - Endpoint registrasi user hanya untuk role `Admin`
   - Endpoint approve disposisi hanya untuk role `Kepala Sekolah`

## Catatan

- Form `POST /surat/upload` harus mengirimkan field `file`, `kategori`, `judul`, `deskripsi`, dan `tujuan_id`.
- Email harus valid pada registrasi, login, OTP, dan reset password.
- Gunakan token `Authorization: Bearer <token>` pada endpoint yang dilindungi.

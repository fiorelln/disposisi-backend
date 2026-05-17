# Audit Program Disposisi Backend

## 1. Ringkasan Umum

Aplikasi backend Go ini menggunakan `gin-gonic` untuk menyediakan fungsi autentikasi, manajemen user dan jabatan, OTP untuk reset password, upload surat PDF, serta alur disposisi surat. Implementasi sudah lengkap dengan middleware otentikasi, migrasi model otomatis, dan dokumentasi API.

## 2. Fitur yang Sudah Selesai

### 2.1 Autentikasi dan Manajemen User

- `POST /auth/login` untuk login pengguna.
- `POST /admin/users` untuk registrasi user baru.
- Password disimpan sebagai hash bcrypt.
- Login mengeluarkan JWT dengan klaim `user_id` dan `roles`.
- Role-user diambil dari `Jabatan.NamaJabatan` dan dibawa dalam token.
- Endpoint registrasi hanya dapat diakses oleh role `Admin`.

### 2.2 OTP dan Reset Password

- `POST /auth/forgot-password` untuk permintaan OTP.
- `POST /auth/verify-otp` untuk verifikasi OTP.
- `POST /auth/reset-password` untuk reset password.
- OTP 6 digit disimpan di tabel `otp` dan kadaluarsa dalam 2 menit.
- Limit 5 OTP dalam 25 menit per user.
- OTP yang sudah terverifikasi hanya berlaku untuk reset password 5 menit berikutnya.

### 2.3 Upload dan Manajemen Surat

- `POST /surat/upload` untuk upload surat PDF.
- `GET /surat/:surat_id` untuk detail surat.
- `GET /surat` untuk daftar surat berdasarkan pengirim atau tujuan.
- Validasi file PDF dan ukuran maksimal 5MB.
- Data surat menyimpan relasi `Pengirim`, `Tujuan`, dan `Disposisi`.

### 2.4 Disposisi Surat

- `POST /disposisi` untuk membuat disposisi.
- `POST /disposisi/approve` untuk approve/reject disposisi.
- `GET /disposisi/surat/:surat_id` untuk mengambil disposisi berdasarkan surat.
- `GET /disposisi` untuk daftar disposisi.
- Endpoint create disposisi hanya dapat diakses oleh role `Admin`, `TU`, `Kepala TU`, dan `Kepala Sekolah`.
- Endpoint approve hanya dapat diakses oleh role `Kepala Sekolah`.

### 2.5 Dashboard Berdasarkan Jabatan

- `GET /dashboard` menampilkan ringkasan berdasarkan role pengguna.
- Menyajikan jumlah surat, disposisi, disposisi yang menunggu, dan approval pending.

## 3. Perbaikan Lanjutan yang Dilakukan

- Menambahkan validasi format email di registrasi, login, OTP, dan reset password.
- Menerapkan `middlewares/RoleMiddleware` pada endpoint `/admin/users`, `/disposisi`, dan `/disposisi/approve`.
- Menambahkan endpoint dashboard `GET /dashboard` untuk ringkasan berdasarkan jabatan.
- Memperkuat validasi input di `controllers/auth.go`.
- Meng-update dokumentasi `README.md` dengan instruksi inisialisasi, menjalankan, dan ringkasan fitur.
- Menyediakan `openapi.yaml` terjemahan Bahasa Indonesia untuk kemudahan developer.

## 4. Rute dan Endpoint Saat Ini

- `POST /auth/login`
- `POST /auth/forgot-password`
- `POST /auth/verify-otp`
- `POST /auth/reset-password`
- `POST /admin/users` (hanya role `Admin`)
- `POST /surat/upload`
- `GET /surat/:surat_id`
- `GET /surat`
- `POST /disposisi`
- `POST /disposisi/approve` (hanya role `Kepala Sekolah`)
- `GET /disposisi/surat/:surat_id`
- `GET /disposisi`

## 5. Kesimpulan

Semua fitur utama untuk alur disposisi surat telah tersedia dan berjalan sesuai desain. Perbaikan keamanan dan validasi telah diterapkan, serta dokumentasi API kini tersedia dalam Bahasa Indonesia untuk mempermudah developer lain.

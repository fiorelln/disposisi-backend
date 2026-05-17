# API Documentation

**Updated**: 17 Mei 2026 - Setelah implementasi perbaikan

## Overview
Dokumentasi lengkap semua API yang terimplementasi di backend `disposisi-backend`. Endpoint sudah menyertakan middleware auth dan alur bisnis disposisi surat.

Base URL: `http://localhost:7000`

---

## 1. Autentikasi dan User Management

### 1.1 POST /auth/login ✅
- Deskripsi: Login user dengan email dan password.
- Autentikasi: Tidak

Request JSON:
```json
{
  "email": "user@example.com",
  "password": "Password123"
}
```

Response sukses (200):
```json
{
  "token": "<jwt_token>",
  "user": {
    "id": 1,
    "name": "Nama User",
    "email": "user@example.com",
    "jabatans": ["Kepala Sekolah", "BK"]
  }
}
```

Response error:
- 400: input JSON invalid
- 401: `Email atau password salah`
- 500: failure on token generation

### 1.2 POST /admin/users ✅
- Deskripsi: Registrasi user baru (admin only).
- Autentikasi: **Bearer token required**

Request JSON:
```json
{
  "name": "Nama User",
  "email": "user@example.com",
  "password": "Passw0rdA",
  "jabatans": [1, 2]
}
```

Response sukses (200):
```json
{
  "message": "User berhasil dibuat"
}
```

---

## 2. Forgot Password dan OTP

### 2.1 POST /auth/forgot-password ✅
- Deskripsi: Minta OTP untuk reset password.
- Autentikasi: Tidak

Response sukses (200):
```json
{
  "message": "OTP berhasil dikirim"
}
```

### 2.2 POST /auth/verify-otp ✅
- Deskripsi: Verifikasi OTP yang dikirim ke email.
- Autentikasi: Tidak

Request JSON:
```json
{
  "email": "user@example.com",
  "otp": "123456"
}
```

Response sukses (200):
```json
{
  "message": "OTP valid"
}
```

### 2.3 POST /auth/reset-password ✅
- Deskripsi: Reset password user dengan OTP verification.
- Autentikasi: Tidak, tetapi OTP harus sudah diverifikasi

Request JSON:
```json
{
  "email": "user@example.com",
  "new_password": "NewPass123"
}
```

Response sukses (200):
```json
{
  "message": "Password berhasil direset"
}
```

**Notes**: 
- Endpoint sekarang mengecek apakah ada OTP yang sudah diverifikasi (`is_used=true`) dalam 5 menit terakhir.
- Jika tidak ada OTP valid, maka error: `OTP belum diverifikasi atau sudah expired`.

---

## 3. Manajemen Surat

### 3.1 POST /surat/upload ✅
- Deskripsi: Upload file PDF surat.
- Autentikasi: **Bearer token required**

Request form-data:
- `file`: PDF file (max 5MB)
- `kategori`: string (required)
- `judul`: string (required)
- `deskripsi`: string (optional)
- `tujuan_id`: uint (required) - ID user yang dituju

Response sukses (200):
```json
{
  "message": "Upload surat berhasil",
  "surat": {
    "id": 1,
    "file_surat": "uploads/surat/1234567890_filename.pdf",
    "status": "dikirim",
    "pengirim_id": 5,
    "tujuan_id": 3,
    "kategori": "Pengumuman",
    "judul": "Judul Surat",
    "deskripsi": "Deskripsi surat"
  }
}
```

Response error:
- 400: `File wajib diupload`
- 400: `File harus PDF`
- 400: `Ukuran file maksimal 5MB`
- 400: `MIME type harus PDF`
- 400: form data tidak valid
- 500: gagal upload file

**Validasi**:
- File extension harus `.pdf`
- Ukuran max 5MB
- MIME type harus `application/pdf`

### 3.2 GET /surat/:surat_id ✅
- Deskripsi: Melihat detail surat.
- Autentikasi: **Bearer token required**

Response sukses (200):
```json
{
  "data": {
    "id": 1,
    "file_surat": "uploads/surat/1234567890_filename.pdf",
    "status": "disetujui",
    "pengirim_id": 5,
    "tujuan_id": 3,
    "kategori": "Pengumuman",
    "judul": "Judul Surat",
    "deskripsi": "Deskripsi surat",
    "pengirim": { "id": 5, "name": "Nama Pengirim", ... },
    "tujuan": { "id": 3, "name": "Nama Tujuan", ... },
    "disposisi": { ... }
  }
}
```

### 3.3 GET /surat (Query parameters) ✅
- Deskripsi: List surat milik user (sebagai pengirim atau tujuan).
- Autentikasi: **Bearer token required**

Query parameters:
- `kategori`: filter by kategori (optional)
- `status`: filter by status (optional)

Response sukses (200):
```json
{
  "data": [
    { "id": 1, "judul": "Surat 1", "status": "dikirim", ... },
    { "id": 2, "judul": "Surat 2", "status": "disetujui", ... }
  ]
}
```

---

## 4. Disposisi Surat

### 4.1 POST /disposisi ✅
- Deskripsi: Membuat disposisi surat (TU ke Kepsek).
- Autentikasi: **Bearer token required**

Request JSON:
```json
{
  "surat_id": 1,
  "tujuan_id": 2,
  "tujuan": "Kepala Sekolah",
  "catatan": "Mohon disetujui"
}
```

Response sukses (200):
```json
{
  "message": "Disposisi berhasil dibuat",
  "disposisi": {
    "id": 1,
    "surat_id": 1,
    "tujuan_id": 2,
    "tujuan": "Kepala Sekolah",
    "catatan": "Mohon disetujui",
    "status": "menunggu",
    "verifikasi_status": "menunggu"
  }
}
```

**Flow**: Status surat otomatis berubah ke `diteruskan`.

### 4.2 POST /disposisi/approve ✅
- Deskripsi: Kepala Sekolah approve/reject surat.
- Autentikasi: **Bearer token required**

Request JSON:
```json
{
  "disposisi_id": 1,
  "is_approved": true,
  "catatan": "Disetujui"
}
```

Response sukses (200):
```json
{
  "message": "Disposisi berhasil diproses",
  "disposisi": {
    "id": 1,
    "surat_id": 1,
    "tujuan_id": 2,
    "verifikator_id": 4,
    "verifikasi_status": "setuju",
    "status": "disetujui"
  }
}
```

**Flow**: 
- Jika `is_approved=true`: status disposisi = `disetujui`, status surat = `disetujui`
- Jika `is_approved=false`: status disposisi = `ditolak`, status surat = `ditolak`

### 4.3 GET /disposisi/surat/:surat_id ✅
- Deskripsi: Melihat disposisi surat tertentu.
- Autentikasi: **Bearer token required**

Response sukses (200):
```json
{
  "data": {
    "id": 1,
    "surat_id": 1,
    "status": "disetujui",
    "verifikasi_status": "setuju",
    "surat": { ... },
    "tujuan_user": { ... },
    "verifikator": { ... }
  }
}
```

### 4.4 GET /disposisi (Query parameters) ✅
- Deskripsi: List disposisi surat (yang dituju atau di-verify oleh user).
- Autentikasi: **Bearer token required**

Query parameters:
- `status`: filter by status (optional) - `menunggu`, `disetujui`, `ditolak`, `selesai`

Response sukses (200):
```json
{
  "data": [
    { "id": 1, "surat_id": 1, "status": "menunggu", ... },
    { "id": 2, "surat_id": 2, "status": "disetujui", ... }
  ]
}
```

---

## 5. Model Database

Tabel yang ada dan dimigrasikan otomatis:
- `users` - User login
- `user_jabatan` - Relasi user dengan jabatan
- `jabatan` - Daftar jabatan
- `otp` - OTP untuk reset password
- `surat` - Surat masuk/keluar
- `disposisi` - Disposisi surat

---

## 6. Perubahan Penting vs Audit Sebelumnya

| Aspek | Sebelum | Sesudah |
|---|---|---|
| Migrasi Database | Hanya User | Semua model |
| Reset Password | Tidak aman (tanpa OTP check) | Aman (OTP required) |
| Upload Surat | Endpoint ada, route kosong | Endpoint aktif & route terdaftar |
| Disposisi | Belum ada | Lengkap dengan create & approve |
| Middleware Auth | Ada code, tidak diterapkan | Diterapkan pada /admin, /surat, /disposisi |
| Validasi File | Hanya ekstensi | Ekstensi + ukuran + MIME type |

---

## 7. Error Handling & Status Code

| Status | Deskripsi |
|---|---|
| 200 | Success |
| 400 | Bad request / validation error |
| 401 | Unauthorized (token invalid/expired) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Resource not found |
| 429 | Too many requests (OTP limit) |
| 500 | Server error |

---

## 8. Rekomendasi untuk Planning Berikutnya

1. **Dashboard per Jabatan**: Buat endpoint `GET /dashboard` yang mengembalikan surat berdasarkan kategori & jabatan user.
2. **Notifikasi**: Tambah WebSocket atau polling untuk notifikasi real-time surat masuk.
3. **Archive & Search**: Implementasikan filter tanggal 3 tahun terakhir seperti di README.
4. **Download Surat**: Endpoint `GET /surat/:surat_id/download` untuk unduh file PDF.
5. **History/Audit Log**: Catat setiap aksi (create, approve, reject) untuk audit trail.
6. **Role-Based Dashboard**: Gunakan `RoleMiddleware` untuk dashboard khusus per role.
7. **Swagger/OpenAPI**: Generate dokumentasi otomatis dengan anotasi Swagger.
8. **Testing**: Tambah unit test dan integration test untuk flow bisnis.


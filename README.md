Aplikasi disposisi surat SMK 2 Singosari adalah sistem internal untuk mengelola surat secara digital agar tidak perlu pengiriman surat secara fisik antar bagian. Dalam sistem ini terdapat lima staff Tata Usaha (TU), di mana dua orang berperan sebagai admin (setara admin utama dan cadangan) yang hanya mengelola akun user, sedangkan tiga lainnya berperan sebagai pengirim surat. User lain terdiri dari berbagai jabatan seperti Kepala TU, Kepala Sekolah, Wakil Kepala Sekolah (Kurikulum, Kesiswaan, Humas, Sarpras), BK (Bimbingan Konseling), BKK, Koordinator, dan Prakerin, dan satu user dapat memiliki lebih dari satu jabatan. Setelah login, setiap user akan masuk ke dashboard sesuai jabatan yang dimilikinya, di mana surat akan masuk berdasarkan kategori atau tujuan surat, misalnya jika surat ditujukan ke BK maka akan masuk ke dashboard BK, dan jika user memiliki lebih dari satu jabatan maka semua surat dari setiap jabatan tersebut akan tampil sesuai kategori masing-masing. Kepala Sekolah berperan memberikan keputusan akhir berupa setuju atau tidak setuju, sementara TU bertugas mengirim surat. Semua surat disimpan berdasarkan kategori, memiliki status, riwayat, dan arsip, serta dapat dicari menggunakan filter hingga tiga tahun terakhir. Sistem juga dilengkapi OTP, reset password, dan reset OTP untuk keamanan, dengan tujuan mempercepat alur surat, mengurangi proses manual, dan memastikan setiap surat terdokumentasi serta mudah ditelusuri berdasarkan kategori dan hak akses masing-masing user.

README NYA AI YAHH YG NYUSUNN 
1. Login

Start

- input email & password
- validasi database users
valid (yes, no)
no, output login gagal, END
yes, app dashboard

END

2. create account

START
input nama, email, password, jabatan
END

3. forgot password
START

input email user
proses validasi email user

valid (yes, no)
no, output email tidak ditemukan, end
yes,
- generate otp
- kirim ke email
- user input otp
- check validasi otp(yes, no)
no, otp salah/expired, END
yes:
- user input password baru
- db update password
- output password berhasil

END

4. surat masuk

START

- Input surat dari luar
- admin (TU) mengirim ke kepsek
- input kepsek mengisi form disposisi
- acc surat (yes, no)

acc suat no:
- update status surat (ditolak)
- surat disimpan 
- END

acc surat yes:
- update status verifikasi (disetujui)
- kirim surat ke user yang dituju
- update status surat (diteruskan)
- surat dibuka
- update status surat (selesai)
- update status alur (selesai)

END

5. surat kelaur

START 

- admin (TU) buat surat keluar
- insert status (menunggu)
- kirim surat ke kepsek
- acc surat (yes, no)

no:
- update status verifikasi (ditolak)
- END

yes:
- update status verifikasi (disetujui)
- notif ke admin (TU) (surat disetujui)
- admin (TU) mengirim ke user
- notifikasi user (menerima surat)
- surat dibaca user
- set tanggal dibaca
- update status (selesai)

END

BIMA TUHH
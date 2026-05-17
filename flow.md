# 1. Login

Start

- input email & password
- validasi database users
valid (yes, no)
no, output login gagal, END
yes, app dashboard

END

## 2. Create account

START
input nama, email, password, jabatan
END

## 3. forgot password

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

## 4. surat masuk

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

## 5. surat kelaur

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
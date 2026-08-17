# Panduan dan Aturan Pengembangan Proyek Caelus Cloud

Dokumen ini memuat standar, aturan, dan tata kelola yang wajib dipatuhi dalam seluruh proses pengembangan sistem Caelus Cloud.

---

## 1. Standar Penulisan Kode (Code Standards)

1. **Pencegahan Duplikasi Kode**
   - Lakukan pemeriksaan menyeluruh sebelum menambahkan fungsi baru untuk memastikan tidak terjadi duplikasi logika atau fungsi ganda (*redundant/duplicate functions*).

2. **Kualitas dan Standar Industri (Enterprise Standard)**
   - Penulisan kode wajib mematuhi konvensi idiomatik dan standar *production-ready* (pola penamaan yang jelas, penanganan error yang eksplisit, dan struktur yang modular).

3. **Dokumentasi Kode (Inline Documentation)**
   - Komentar hanya diperbolehkan pada tingkat fungsi (*function docstring/signature comments*).
   - Format komentar wajib berupa penjelasan teknis mengenai fungsionalitas, parameter, dan nilai kembalian fungsi.
   - Dilarang menulis komentar bergaya tutorial atau naratif informal.
   - Dilarang memberikan komentar pada baris logika internal di luar deklarasi fungsi.

4. **Verifikasi Sebelum Implementasi**
   - Lakukan analisis kebutuhan, verifikasi dependensi, dan rancangan logika sebelum melakukan eksekusi penulisan kode.

---

## 2. Standar Dokumentasi & Pelacakan Proyek

1. **Format Bebas Emoji**
   - Seluruh dokumen teknis bertipe Markdown (`*.md`) dilarang memuat karakter emoji dalam bentuk apa pun agar mematuhi standar formal enterprise.

2. **Pelacakan Progres (CHANGELOG.md)**
   - Setiap perubahan fitur, perbaikan bug, atau penyesuaian arsitektur wajib dicatat secara berkala dan konsisten pada file `CHANGELOG.md`.

3. **Integritas Dokumentasi (Anti-Overclaim)**
   - Isi file `README.md`, status modul, dan dokumentasi pendukung wajib mencerminkan implementasi yang sudah selesai dikerjakan secara riil dan terverifikasi.
   - Dilarang mencantumkan klaim fitur yang belum diimplementasikan sebagai fitur yang sudah selesai.

---

## 3. Alur Kerja dan Arsitektur (Development Flow & Architecture)

1. **Alur Kerja Modular Berurutan**
   - Pengerjaan dilakukan secara bertahap per modul (*modular workflow*).
   - Setiap modul harus diselesaikan dan diverifikasi sebelum melanjutkan ke modul berikutnya, tanpa melangkahi urutan tahapan yang telah ditetapkan.

2. **Penerapan Clean Architecture**
   - Struktur kode dan dependensi wajib mematuhi prinsip *Clean Architecture* (pemisahan domain entity, use cases/application logic, interface adapters/repositories, dan external frameworks/drivers).
   - Dependensi harus selalu mengarah ke dalam (*dependency rule*), di mana domain logic tidak boleh bergantung pada implementasi database, framework HTTP, atau driver eksternal.
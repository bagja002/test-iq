UPDATE test_configs
SET question_count = 3
WHERE test_type = 'SKB' AND is_active = TRUE;

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebagai Manajer Kopreasi (KDMP), langkah pertama saat menemukan selisih antara saldo kas dan buku kas harian adalah?', 'medium', 'SKB', 'MANAJER_KOPERASI_KDMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Melakukan rekonsiliasi dan menelusuri bukti transaksi sebelum mengambil keputusan', TRUE),
    (LAST_INSERT_ID(), 'B', 'Langsung menutup operasional koperasi selama seminggu', FALSE),
    (LAST_INSERT_ID(), 'C', 'Menghapus catatan lama agar saldo kembali sama', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menunggu audit tahunan tanpa tindak lanjut', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Indikator yang paling tepat untuk menilai kesehatan usaha koperasi simpan pinjam adalah?', 'medium', 'SKB', 'MANAJER_KOPERASI_KDMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Jumlah rapat internal yang dilakukan setiap bulan', FALSE),
    (LAST_INSERT_ID(), 'B', 'Rasio kredit bermasalah, likuiditas, dan pertumbuhan anggota aktif', TRUE),
    (LAST_INSERT_ID(), 'C', 'Jumlah spanduk promosi di kantor koperasi', FALSE),
    (LAST_INSERT_ID(), 'D', 'Banyaknya meja pelayanan yang tersedia', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Ketika anggota koperasi mengeluhkan proses pinjaman yang lambat, tindakan manajerial terbaik adalah?', 'easy', 'SKB', 'MANAJER_KOPERASI_KDMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Menyalahkan petugas front office di depan anggota', FALSE),
    (LAST_INSERT_ID(), 'B', 'Meninjau alur layanan, mengukur bottleneck, lalu memperbaiki SOP proses pinjaman', TRUE),
    (LAST_INSERT_ID(), 'C', 'Menghentikan sementara semua pengajuan pinjaman', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menghapus persyaratan administrasi tanpa kajian risiko', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebagai Manager Operesial (KNMP), prioritas utama saat target distribusi harian terancam tidak tercapai adalah?', 'medium', 'SKB', 'MANAGER_OPERASIONAL_KNMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Mengumpulkan data hambatan operasional lalu menyesuaikan sumber daya dan jadwal pengiriman', TRUE),
    (LAST_INSERT_ID(), 'B', 'Menunggu laporan akhir hari tanpa intervensi', FALSE),
    (LAST_INSERT_ID(), 'C', 'Mengurangi jumlah pelanggan yang dilayani tanpa pemberitahuan', FALSE),
    (LAST_INSERT_ID(), 'D', 'Memindahkan seluruh staf ke fungsi administrasi', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Metode paling efektif untuk menjaga konsistensi proses operasional di beberapa shift adalah?', 'easy', 'SKB', 'MANAGER_OPERASIONAL_KNMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Mengandalkan kebiasaan masing-masing supervisor', FALSE),
    (LAST_INSERT_ID(), 'B', 'Menggunakan SOP baku, briefing shift, dan monitoring KPI operasional', TRUE),
    (LAST_INSERT_ID(), 'C', 'Membebaskan tiap shift membuat aturan sendiri', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menghapus target kinerja harian', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Jika terjadi lonjakan order mendadak, keputusan operasional yang paling tepat adalah?', 'medium', 'SKB', 'MANAGER_OPERASIONAL_KNMP', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Menolak semua order tambahan agar tim tidak lelah', FALSE),
    (LAST_INSERT_ID(), 'B', 'Mengatur ulang kapasitas, memprioritaskan order kritis, dan berkoordinasi lintas fungsi', TRUE),
    (LAST_INSERT_ID(), 'C', 'Melepas pengawasan proses agar pekerjaan lebih cepat', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menghentikan pengiriman yang sudah dijadwalkan', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebagai Kepala Produksi (KNMP), tindakan pertama saat output turun karena mesin sering berhenti adalah?', 'medium', 'SKB', 'KEPALA_PRODUKSI', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Menambah target produksi tanpa analisis gangguan', FALSE),
    (LAST_INSERT_ID(), 'B', 'Mencatat downtime, mencari akar masalah, dan menjadwalkan perbaikan terencana', TRUE),
    (LAST_INSERT_ID(), 'C', 'Mengganti semua operator pada hari yang sama', FALSE),
    (LAST_INSERT_ID(), 'D', 'Mengabaikan gangguan kecil selama produk masih jadi', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Ukuran keberhasilan produksi yang paling relevan untuk dipantau harian adalah?', 'easy', 'SKB', 'KEPALA_PRODUKSI', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Jumlah rapat yang dilakukan kepala produksi', FALSE),
    (LAST_INSERT_ID(), 'B', 'Output aktual, tingkat reject, efisiensi mesin, dan ketepatan jadwal', TRUE),
    (LAST_INSERT_ID(), 'C', 'Jumlah tamu yang datang ke area produksi', FALSE),
    (LAST_INSERT_ID(), 'D', 'Banyaknya pesan di grup kerja', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Saat bahan baku datang dengan kualitas tidak konsisten, langkah terbaik kepala produksi adalah?', 'medium', 'SKB', 'KEPALA_PRODUKSI', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Tetap memakai semua bahan agar target tercapai', FALSE),
    (LAST_INSERT_ID(), 'B', 'Berkoordinasi dengan QC dan supplier untuk segregasi material serta tindakan korektif', TRUE),
    (LAST_INSERT_ID(), 'C', 'Mencampur semua bahan agar perbedaan tidak terlihat', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menghentikan pencatatan penerimaan bahan', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebagai Pengelola Keuangan (KNMP), dokumen yang paling penting saat melakukan verifikasi pembayaran vendor adalah?', 'easy', 'SKB', 'PENGELOLA_KEUANGAN', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Faktur, purchase order, bukti penerimaan barang/jasa, dan otorisasi pembayaran', TRUE),
    (LAST_INSERT_ID(), 'B', 'Catatan percakapan informal dengan vendor saja', FALSE),
    (LAST_INSERT_ID(), 'C', 'Foto gudang tanpa dokumen pendukung', FALSE),
    (LAST_INSERT_ID(), 'D', 'Brosur perusahaan vendor', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Tujuan utama rekonsiliasi bank bulanan adalah?', 'easy', 'SKB', 'PENGELOLA_KEUANGAN', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Menambah saldo rekening secara administratif', FALSE),
    (LAST_INSERT_ID(), 'B', 'Memastikan catatan perusahaan sesuai dengan mutasi bank dan menemukan selisih', TRUE),
    (LAST_INSERT_ID(), 'C', 'Menghapus transaksi kecil yang sulit dilacak', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menunda pencatatan transaksi sampai audit selesai', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Jika realisasi biaya operasional melebihi anggaran, langkah paling tepat bagi pengelola keuangan adalah?', 'medium', 'SKB', 'PENGELOLA_KEUANGAN', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Menyusun analisis varian, mengidentifikasi penyebab, dan melaporkan rekomendasi pengendalian biaya', TRUE),
    (LAST_INSERT_ID(), 'B', 'Mengubah angka anggaran agar sesuai realisasi tanpa persetujuan', FALSE),
    (LAST_INSERT_ID(), 'C', 'Menghapus sebagian bukti biaya', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menghentikan semua pembayaran mendadak', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebagai Penjamin Mutu (KNMP), langkah awal ketika ditemukan produk tidak sesuai spesifikasi adalah?', 'medium', 'SKB', 'PENJAMIN_MUTU', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Memisahkan produk, mencatat temuan, dan menjalankan prosedur penanganan ketidaksesuaian', TRUE),
    (LAST_INSERT_ID(), 'B', 'Melepas produk ke pelanggan agar gudang tidak penuh', FALSE),
    (LAST_INSERT_ID(), 'C', 'Menghapus catatan inspeksi sebelumnya', FALSE),
    (LAST_INSERT_ID(), 'D', 'Langsung menyalahkan operator tanpa investigasi', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Tujuan utama audit mutu internal adalah?', 'easy', 'SKB', 'PENJAMIN_MUTU', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Mencari kesalahan personal untuk diberikan sanksi', FALSE),
    (LAST_INSERT_ID(), 'B', 'Memastikan proses sesuai standar dan menemukan peluang perbaikan sistem', TRUE),
    (LAST_INSERT_ID(), 'C', 'Mengganti semua dokumen mutu setiap bulan', FALSE),
    (LAST_INSERT_ID(), 'D', 'Menurunkan target kualitas agar mudah tercapai', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Dokumen yang paling penting untuk melacak tindakan korektif dari temuan mutu adalah?', 'medium', 'SKB', 'PENJAMIN_MUTU', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Daftar hadir rapat mingguan saja', FALSE),
    (LAST_INSERT_ID(), 'B', 'Log CAPA yang mencatat akar masalah, tindakan, PIC, dan status penyelesaian', TRUE),
    (LAST_INSERT_ID(), 'C', 'Poster motivasi di area kerja', FALSE),
    (LAST_INSERT_ID(), 'D', 'Catatan pribadi supervisor tanpa format baku', FALSE);

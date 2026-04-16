ALTER TABLE questions
    ADD COLUMN question_index VARCHAR(16) NULL AFTER difficulty,
    ADD COLUMN subtest_code VARCHAR(64) NULL AFTER question_index;

ALTER TABLE attempt_questions
    ADD COLUMN question_index VARCHAR(16) NULL AFTER question_id,
    ADD COLUMN subtest_code VARCHAR(64) NULL AFTER question_index;

CREATE INDEX idx_questions_question_index ON questions (question_index);
CREATE INDEX idx_questions_subtest_code ON questions (subtest_code);
CREATE INDEX idx_attempt_questions_question_index ON attempt_questions (question_index);
CREATE INDEX idx_attempt_questions_subtest_code ON attempt_questions (subtest_code);

UPDATE test_configs SET is_active = FALSE WHERE is_active = TRUE;

INSERT INTO test_configs (title, duration_minutes, question_count, is_active)
VALUES ('Test IQ Screening 15 Subtes', 35, 15, TRUE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Apa persamaan paling tepat antara dokter dan guru?', 'easy', 'VCI', 'SIMILARITIES', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Keduanya profesi yang membantu dan melayani orang lain', TRUE),
    (LAST_INSERT_ID(), 'B', 'Keduanya selalu bekerja malam hari', FALSE),
    (LAST_INSERT_ID(), 'C', 'Keduanya hanya bekerja di rumah sakit', FALSE),
    (LAST_INSERT_ID(), 'D', 'Keduanya menjual barang yang sama', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Kata adaptif paling dekat artinya dengan apa?', 'easy', 'VCI', 'VOCABULARY', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Mampu menyesuaikan diri', TRUE),
    (LAST_INSERT_ID(), 'B', 'Tidak mau berubah', FALSE),
    (LAST_INSERT_ID(), 'C', 'Sangat lambat', FALSE),
    (LAST_INSERT_ID(), 'D', 'Sulit memahami', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Organ tubuh yang berfungsi memompa darah adalah?', 'easy', 'VCI', 'INFORMATION', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Paru-paru', FALSE),
    (LAST_INSERT_ID(), 'B', 'Jantung', TRUE),
    (LAST_INSERT_ID(), 'C', 'Ginjal', FALSE),
    (LAST_INSERT_ID(), 'D', 'Lambung', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Mengapa perusahaan memiliki SOP keselamatan kerja?', 'medium', 'VCI', 'COMPREHENSION', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Agar pekerjaan terasa lebih lama', FALSE),
    (LAST_INSERT_ID(), 'B', 'Untuk mengurangi risiko dan menjaga konsistensi kerja', TRUE),
    (LAST_INSERT_ID(), 'C', 'Agar biaya operasional selalu bertambah', FALSE),
    (LAST_INSERT_ID(), 'D', 'Supaya setiap divisi bebas membuat aturan sendiri', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Pola target 2x2 adalah Baris 1 Merah Biru dan Baris 2 Merah Biru. Pilih susunan yang identik.', 'medium', 'PRI', 'BLOCK_DESIGN', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Baris 1 Merah Biru dan Baris 2 Merah Biru', TRUE),
    (LAST_INSERT_ID(), 'B', 'Baris 1 Biru Merah dan Baris 2 Merah Biru', FALSE),
    (LAST_INSERT_ID(), 'C', 'Baris 1 Merah Merah dan Baris 2 Biru Biru', FALSE),
    (LAST_INSERT_ID(), 'D', 'Baris 1 Biru Biru dan Baris 2 Merah Merah', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Lengkapi pola berikut: Baris 1 segitiga lingkaran segitiga, Baris 2 lingkaran segitiga ?, Baris 3 segitiga lingkaran segitiga.', 'medium', 'PRI', 'MATRIX_REASONING', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Segitiga', FALSE),
    (LAST_INSERT_ID(), 'B', 'Lingkaran', TRUE),
    (LAST_INSERT_ID(), 'C', 'Persegi', FALSE),
    (LAST_INSERT_ID(), 'D', 'Bintang', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Jika targetnya persegi besar yang dibentuk dari empat potongan segitiga, kombinasi bagian mana yang paling tepat?', 'medium', 'PRI', 'VISUAL_PUZZLES', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'P1 ditambah P2 ditambah P3 ditambah P4', TRUE),
    (LAST_INSERT_ID(), 'B', 'P1 ditambah P2 ditambah P5', FALSE),
    (LAST_INSERT_ID(), 'C', 'P2 ditambah P3 ditambah P5', FALSE),
    (LAST_INSERT_ID(), 'D', 'P1 ditambah P4 ditambah P5', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebuah gambar jam analog kehilangan bagian penting. Bagian mana yang paling penting untuk melengkapi gambar?', 'easy', 'PRI', 'PICTURE_COMPLETION', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Jarum jam', TRUE),
    (LAST_INSERT_ID(), 'B', 'Bingkai dekoratif', FALSE),
    (LAST_INSERT_ID(), 'C', 'Warna latar', FALSE),
    (LAST_INSERT_ID(), 'D', 'Bayangan meja', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Jika satu persegi setara dengan dua lingkaran, maka dua persegi setara dengan berapa lingkaran?', 'easy', 'PRI', 'FIGURE_WEIGHTS', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '2 lingkaran', FALSE),
    (LAST_INSERT_ID(), 'B', '3 lingkaran', FALSE),
    (LAST_INSERT_ID(), 'C', '4 lingkaran', TRUE),
    (LAST_INSERT_ID(), 'D', '5 lingkaran', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Deret angka 7 2 9 4. Jika diminta mengulang dari belakang, jawaban yang benar adalah?', 'medium', 'WMI', 'DIGIT_SPAN', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '4 9 2 7', TRUE),
    (LAST_INSERT_ID(), 'B', '7 2 9 4', FALSE),
    (LAST_INSERT_ID(), 'C', '9 4 2 7', FALSE),
    (LAST_INSERT_ID(), 'D', '4 7 9 2', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Sebuah gudang memiliki 48 kotak. Sebanyak 15 terjual dan 9 rusak. Berapa sisa kotak yang masih layak?', 'medium', 'WMI', 'ARITHMETIC', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '22', FALSE),
    (LAST_INSERT_ID(), 'B', '23', FALSE),
    (LAST_INSERT_ID(), 'C', '24', TRUE),
    (LAST_INSERT_ID(), 'D', '25', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Urutkan B 3 A 1 dengan aturan angka dari kecil ke besar lalu huruf alfabetis. Jawaban yang benar adalah?', 'medium', 'WMI', 'LETTER_NUMBER_SEQUENCING', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '1 3 A B', TRUE),
    (LAST_INSERT_ID(), 'B', 'A B 1 3', FALSE),
    (LAST_INSERT_ID(), 'C', '1 A 3 B', FALSE),
    (LAST_INSERT_ID(), 'D', '3 1 B A', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Target simbol adalah persen dan at. Pada kelompok simbol pagar at seru persen dolar, pernyataan yang benar adalah?', 'easy', 'PSI', 'SYMBOL_SEARCH', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', 'Kedua target ada', TRUE),
    (LAST_INSERT_ID(), 'B', 'Hanya persen yang ada', FALSE),
    (LAST_INSERT_ID(), 'C', 'Hanya at yang ada', FALSE),
    (LAST_INSERT_ID(), 'D', 'Tidak ada target', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Kode simbol adalah 1 segitiga, 2 lingkaran, 3 persegi, 4 bintang. Simbol lingkaran persegi segitiga diterjemahkan menjadi?', 'medium', 'PSI', 'CODING', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '2 3 1', TRUE),
    (LAST_INSERT_ID(), 'B', '1 2 3', FALSE),
    (LAST_INSERT_ID(), 'C', '3 2 1', FALSE),
    (LAST_INSERT_ID(), 'D', '2 1 4', FALSE);

INSERT INTO questions (prompt, difficulty, question_index, subtest_code, status)
VALUES ('Target yang harus ditandai adalah A dan 7. Pada deret B 7 A 3 A 9 7 C, berapa jumlah target yang benar?', 'easy', 'PSI', 'CANCELLATION', 'PUBLISHED');
INSERT INTO question_options (question_id, `key`, content, is_correct) VALUES
    (LAST_INSERT_ID(), 'A', '3', FALSE),
    (LAST_INSERT_ID(), 'B', '4', TRUE),
    (LAST_INSERT_ID(), 'C', '5', FALSE),
    (LAST_INSERT_ID(), 'D', '6', FALSE);

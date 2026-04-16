ALTER TABLE questions
    ADD COLUMN prompt_media_url VARCHAR(255) NULL AFTER prompt,
    ADD COLUMN prompt_media_alt VARCHAR(255) NULL AFTER prompt_media_url;

ALTER TABLE question_options
    ADD COLUMN media_url VARCHAR(255) NULL AFTER content,
    ADD COLUMN media_alt VARCHAR(255) NULL AFTER media_url;

ALTER TABLE attempt_questions
    ADD COLUMN prompt_media_url VARCHAR(255) NULL AFTER prompt_snapshot,
    ADD COLUMN prompt_media_alt VARCHAR(255) NULL AFTER prompt_media_url;

UPDATE questions
SET prompt = 'Perhatikan gambar target. Pilih susunan balok yang identik.',
    prompt_media_url = '/question-assets/pri/block-design/prompt.svg',
    prompt_media_alt = 'Target block design empat kotak dengan pola warna merah dan biru'
WHERE subtest_code = 'BLOCK_DESIGN'
  AND prompt = 'Pola target 2x2 adalah Baris 1 Merah Biru dan Baris 2 Merah Biru. Pilih susunan yang identik.';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan A',
    qo.media_url = '/question-assets/pri/block-design/option-a.svg',
    qo.media_alt = 'Pilihan A susunan balok identik dengan target'
WHERE q.subtest_code = 'BLOCK_DESIGN' AND q.prompt = 'Perhatikan gambar target. Pilih susunan balok yang identik.' AND qo.`key` = 'A';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan B',
    qo.media_url = '/question-assets/pri/block-design/option-b.svg',
    qo.media_alt = 'Pilihan B susunan balok berbeda pada baris atas'
WHERE q.subtest_code = 'BLOCK_DESIGN' AND q.prompt = 'Perhatikan gambar target. Pilih susunan balok yang identik.' AND qo.`key` = 'B';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan C',
    qo.media_url = '/question-assets/pri/block-design/option-c.svg',
    qo.media_alt = 'Pilihan C susunan balok berkelompok per warna'
WHERE q.subtest_code = 'BLOCK_DESIGN' AND q.prompt = 'Perhatikan gambar target. Pilih susunan balok yang identik.' AND qo.`key` = 'C';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan D',
    qo.media_url = '/question-assets/pri/block-design/option-d.svg',
    qo.media_alt = 'Pilihan D susunan balok dominan biru di baris atas'
WHERE q.subtest_code = 'BLOCK_DESIGN' AND q.prompt = 'Perhatikan gambar target. Pilih susunan balok yang identik.' AND qo.`key` = 'D';

UPDATE questions
SET prompt = 'Lengkapi pola visual pada kotak yang kosong.',
    prompt_media_url = '/question-assets/pri/matrix-reasoning/prompt.svg',
    prompt_media_alt = 'Matriks visual tiga kali tiga dengan satu kotak kosong di tengah bawah'
WHERE subtest_code = 'MATRIX_REASONING'
  AND prompt = 'Lengkapi pola berikut: Baris 1 segitiga lingkaran segitiga, Baris 2 lingkaran segitiga ?, Baris 3 segitiga lingkaran segitiga.';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan A',
    qo.media_url = '/question-assets/pri/matrix-reasoning/option-a.svg',
    qo.media_alt = 'Pilihan A berupa segitiga'
WHERE q.subtest_code = 'MATRIX_REASONING' AND q.prompt = 'Lengkapi pola visual pada kotak yang kosong.' AND qo.`key` = 'A';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan B',
    qo.media_url = '/question-assets/pri/matrix-reasoning/option-b.svg',
    qo.media_alt = 'Pilihan B berupa lingkaran'
WHERE q.subtest_code = 'MATRIX_REASONING' AND q.prompt = 'Lengkapi pola visual pada kotak yang kosong.' AND qo.`key` = 'B';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan C',
    qo.media_url = '/question-assets/pri/matrix-reasoning/option-c.svg',
    qo.media_alt = 'Pilihan C berupa persegi'
WHERE q.subtest_code = 'MATRIX_REASONING' AND q.prompt = 'Lengkapi pola visual pada kotak yang kosong.' AND qo.`key` = 'C';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan D',
    qo.media_url = '/question-assets/pri/matrix-reasoning/option-d.svg',
    qo.media_alt = 'Pilihan D berupa bintang'
WHERE q.subtest_code = 'MATRIX_REASONING' AND q.prompt = 'Lengkapi pola visual pada kotak yang kosong.' AND qo.`key` = 'D';

UPDATE questions
SET prompt = 'Perhatikan bentuk target. Pilih potongan yang dapat membentuk gambar tersebut.',
    prompt_media_url = '/question-assets/pri/visual-puzzles/prompt.svg',
    prompt_media_alt = 'Bentuk target berupa rumah sederhana'
WHERE subtest_code = 'VISUAL_PUZZLES'
  AND prompt = 'Jika targetnya persegi besar yang dibentuk dari empat potongan segitiga, kombinasi bagian mana yang paling tepat?';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan A',
    qo.media_url = '/question-assets/pri/visual-puzzles/option-a.svg',
    qo.media_alt = 'Pilihan A empat potongan yang dapat membentuk rumah'
WHERE q.subtest_code = 'VISUAL_PUZZLES' AND q.prompt = 'Perhatikan bentuk target. Pilih potongan yang dapat membentuk gambar tersebut.' AND qo.`key` = 'A';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan B',
    qo.media_url = '/question-assets/pri/visual-puzzles/option-b.svg',
    qo.media_alt = 'Pilihan B tiga potongan yang tidak lengkap'
WHERE q.subtest_code = 'VISUAL_PUZZLES' AND q.prompt = 'Perhatikan bentuk target. Pilih potongan yang dapat membentuk gambar tersebut.' AND qo.`key` = 'B';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan C',
    qo.media_url = '/question-assets/pri/visual-puzzles/option-c.svg',
    qo.media_alt = 'Pilihan C kombinasi potongan yang menyisakan celah'
WHERE q.subtest_code = 'VISUAL_PUZZLES' AND q.prompt = 'Perhatikan bentuk target. Pilih potongan yang dapat membentuk gambar tersebut.' AND qo.`key` = 'C';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan D',
    qo.media_url = '/question-assets/pri/visual-puzzles/option-d.svg',
    qo.media_alt = 'Pilihan D kombinasi potongan yang bentuknya tidak cocok'
WHERE q.subtest_code = 'VISUAL_PUZZLES' AND q.prompt = 'Perhatikan bentuk target. Pilih potongan yang dapat membentuk gambar tersebut.' AND qo.`key` = 'D';

UPDATE questions
SET prompt = 'Perhatikan gambar objek. Pilih bagian penting yang hilang.',
    prompt_media_url = '/question-assets/pri/picture-completion/prompt.svg',
    prompt_media_alt = 'Jam analog tanpa jarum penunjuk waktu'
WHERE subtest_code = 'PICTURE_COMPLETION'
  AND prompt = 'Sebuah gambar jam analog kehilangan bagian penting. Bagian mana yang paling penting untuk melengkapi gambar?';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan A',
    qo.media_url = '/question-assets/pri/picture-completion/option-a.svg',
    qo.media_alt = 'Pilihan A berupa jarum jam'
WHERE q.subtest_code = 'PICTURE_COMPLETION' AND q.prompt = 'Perhatikan gambar objek. Pilih bagian penting yang hilang.' AND qo.`key` = 'A';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan B',
    qo.media_url = '/question-assets/pri/picture-completion/option-b.svg',
    qo.media_alt = 'Pilihan B berupa bingkai dekoratif'
WHERE q.subtest_code = 'PICTURE_COMPLETION' AND q.prompt = 'Perhatikan gambar objek. Pilih bagian penting yang hilang.' AND qo.`key` = 'B';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan C',
    qo.media_url = '/question-assets/pri/picture-completion/option-c.svg',
    qo.media_alt = 'Pilihan C berupa latar bulat'
WHERE q.subtest_code = 'PICTURE_COMPLETION' AND q.prompt = 'Perhatikan gambar objek. Pilih bagian penting yang hilang.' AND qo.`key` = 'C';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan D',
    qo.media_url = '/question-assets/pri/picture-completion/option-d.svg',
    qo.media_alt = 'Pilihan D berupa bayangan'
WHERE q.subtest_code = 'PICTURE_COMPLETION' AND q.prompt = 'Perhatikan gambar objek. Pilih bagian penting yang hilang.' AND qo.`key` = 'D';

UPDATE questions
SET prompt = 'Perhatikan keseimbangan bentuk pada gambar. Pilih jawaban yang setara.',
    prompt_media_url = '/question-assets/pri/figure-weights/prompt.svg',
    prompt_media_alt = 'Satu persegi seimbang dengan dua lingkaran pada timbangan'
WHERE subtest_code = 'FIGURE_WEIGHTS'
  AND prompt = 'Jika satu persegi setara dengan dua lingkaran, maka dua persegi setara dengan berapa lingkaran?';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan A',
    qo.media_url = '/question-assets/pri/figure-weights/option-a.svg',
    qo.media_alt = 'Pilihan A berisi dua lingkaran'
WHERE q.subtest_code = 'FIGURE_WEIGHTS' AND q.prompt = 'Perhatikan keseimbangan bentuk pada gambar. Pilih jawaban yang setara.' AND qo.`key` = 'A';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan B',
    qo.media_url = '/question-assets/pri/figure-weights/option-b.svg',
    qo.media_alt = 'Pilihan B berisi tiga lingkaran'
WHERE q.subtest_code = 'FIGURE_WEIGHTS' AND q.prompt = 'Perhatikan keseimbangan bentuk pada gambar. Pilih jawaban yang setara.' AND qo.`key` = 'B';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan C',
    qo.media_url = '/question-assets/pri/figure-weights/option-c.svg',
    qo.media_alt = 'Pilihan C berisi empat lingkaran'
WHERE q.subtest_code = 'FIGURE_WEIGHTS' AND q.prompt = 'Perhatikan keseimbangan bentuk pada gambar. Pilih jawaban yang setara.' AND qo.`key` = 'C';

UPDATE question_options qo
JOIN questions q ON q.id = qo.question_id
SET qo.content = 'Pilihan D',
    qo.media_url = '/question-assets/pri/figure-weights/option-d.svg',
    qo.media_alt = 'Pilihan D berisi lima lingkaran'
WHERE q.subtest_code = 'FIGURE_WEIGHTS' AND q.prompt = 'Perhatikan keseimbangan bentuk pada gambar. Pilih jawaban yang setara.' AND qo.`key` = 'D';

UPDATE questions
SET prompt = 'Perhatikan target simbol pada gambar. Pilih pernyataan yang benar.',
    prompt_media_url = '/question-assets/psi/symbol-search/prompt.svg',
    prompt_media_alt = 'Target simbol persen dan at dengan deret simbol pencarian'
WHERE subtest_code = 'SYMBOL_SEARCH'
  AND prompt = 'Target simbol adalah persen dan at. Pada kelompok simbol pagar at seru persen dolar, pernyataan yang benar adalah?';

UPDATE questions
SET prompt = 'Perhatikan legenda simbol pada gambar. Ubah deret simbol menjadi angka.',
    prompt_media_url = '/question-assets/psi/coding/prompt.svg',
    prompt_media_alt = 'Legenda simbol segitiga lingkaran persegi bintang dengan angka satu sampai empat'
WHERE subtest_code = 'CODING'
  AND prompt = 'Kode simbol adalah 1 segitiga, 2 lingkaran, 3 persegi, 4 bintang. Simbol lingkaran persegi segitiga diterjemahkan menjadi?';

UPDATE questions
SET prompt = 'Perhatikan deret pada gambar. Hitung jumlah target yang harus ditandai.',
    prompt_media_url = '/question-assets/psi/cancellation/prompt.svg',
    prompt_media_alt = 'Target A dan angka tujuh dengan deret simbol B 7 A 3 A 9 7 C'
WHERE subtest_code = 'CANCELLATION'
  AND prompt = 'Target yang harus ditandai adalah A dan 7. Pada deret B 7 A 3 A 9 7 C, berapa jumlah target yang benar?';

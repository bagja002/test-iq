ALTER TABLE users
  ADD COLUMN position VARCHAR(191) NOT NULL DEFAULT '' AFTER name;

UPDATE users
SET position = 'Belum diisi'
WHERE position IS NULL OR TRIM(position) = '';

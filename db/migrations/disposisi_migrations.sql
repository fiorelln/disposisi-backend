-- ============================================
-- Disposisi System Database Migrations
-- ============================================
-- Jalankan migrations ini untuk setup database
-- Gunakan: migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/db" up

-- Migration: 000001_alter_disposisi_table.up.sql
-- Purpose: Add parent-child relationship support untuk chain forwarding

BEGIN;

-- 1. Add new columns untuk parent-child relationship
ALTER TABLE disposisi 
ADD COLUMN IF NOT EXISTS parent_disposisi_id BIGINT;

ALTER TABLE disposisi 
ADD COLUMN IF NOT EXISTS level INT DEFAULT 0;

-- 2. Add timestamps untuk tracking
ALTER TABLE disposisi 
ADD COLUMN IF NOT EXISTS baca_at TIMESTAMP;

ALTER TABLE disposisi 
ADD COLUMN IF NOT EXISTS complete_at TIMESTAMP;

-- 3. Add soft delete untuk audit trail
ALTER TABLE disposisi 
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- 4. Rename atau add columns untuk consistency
-- Jika column lama masih ada, bisa ditambahkan mapping
-- ALTER TABLE disposisi RENAME COLUMN id_disposisi TO id; -- Hanya jika using migrate tool

-- 5. Add foreign key constraints
ALTER TABLE disposisi 
ADD CONSTRAINT fk_disposisi_parent 
FOREIGN KEY (parent_disposisi_id) 
REFERENCES disposisi(id) 
ON DELETE SET NULL;

-- 6. Create indexes untuk performance
CREATE INDEX IF NOT EXISTS idx_disposisi_to_user_id 
ON disposisi(to_user_id, deleted_at);

CREATE INDEX IF NOT EXISTS idx_disposisi_from_user_id 
ON disposisi(from_user_id, deleted_at);

CREATE INDEX IF NOT EXISTS idx_disposisi_surat_masuk_id 
ON disposisi(surat_masuk_id, deleted_at);

CREATE INDEX IF NOT EXISTS idx_disposisi_parent_id 
ON disposisi(parent_disposisi_id);

CREATE INDEX IF NOT EXISTS idx_disposisi_status 
ON disposisi(status, deleted_at);

CREATE INDEX IF NOT EXISTS idx_disposisi_level 
ON disposisi(level);

CREATE INDEX IF NOT EXISTS idx_disposisi_created_at 
ON disposisi(created_at DESC, deleted_at);

-- 7. Composite indexes untuk common queries
CREATE INDEX IF NOT EXISTS idx_disposisi_inbox 
ON disposisi(to_user_id, deleted_at, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_disposisi_sent 
ON disposisi(from_user_id, deleted_at, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_disposisi_tree 
ON disposisi(surat_masuk_id, level, deleted_at);

COMMIT;


-- ============================================
-- Migration: 000002_create_notifikasi_table.up.sql
-- Purpose: Create notification table untuk tracking disposisi events

BEGIN;

CREATE TABLE IF NOT EXISTS notifikasi (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id_user) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- disposisi_received, disposisi_completed, disposisi_rejected
    title VARCHAR(255) NOT NULL,
    message TEXT,
    disposisi_id BIGINT REFERENCES disposisi(id) ON DELETE SET NULL,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes untuk notifikasi
CREATE INDEX IF NOT EXISTS idx_notifikasi_user_id 
ON notifikasi(user_id, deleted_at);

CREATE INDEX IF NOT EXISTS idx_notifikasi_unread 
ON notifikasi(user_id, is_read, deleted_at);

CREATE INDEX IF NOT EXISTS idx_notifikasi_created_at 
ON notifikasi(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifikasi_disposisi_id 
ON notifikasi(disposisi_id);

COMMIT;


-- ============================================
-- Migration: 000003_seed_jabatan_data.up.sql
-- Purpose: Seed role/jabatan data untuk permission system

BEGIN;

INSERT INTO jabatan (nama_jabatan, level_akses) VALUES
    ('KEPALA_SEKOLAH', 'admin'),
    ('WAKIL_KEPALA', 'manager'),
    ('WAKIL_KURIKULUM', 'manager'),
    ('WAKIL_KESISWAAN', 'manager'),
    ('WAKIL_SARANA', 'manager'),
    ('TATA_USAHA', 'operator'),
    ('GURU', 'user'),
    ('STAFF', 'user'),
    ('KEPALA_PERPUS', 'manager'),
    ('KEPALA_LAB', 'manager'),
    ('BK', 'manager')
ON CONFLICT DO NOTHING;

COMMIT;


-- ============================================
-- Migration Rollback Scripts
-- ============================================

-- Migration: 000001_alter_disposisi_table.down.sql
BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS idx_disposisi_inbox;
DROP INDEX IF EXISTS idx_disposisi_sent;
DROP INDEX IF EXISTS idx_disposisi_tree;
DROP INDEX IF EXISTS idx_disposisi_created_at;
DROP INDEX IF EXISTS idx_disposisi_level;
DROP INDEX IF EXISTS idx_disposisi_status;
DROP INDEX IF EXISTS idx_disposisi_parent_id;
DROP INDEX IF EXISTS idx_disposisi_surat_masuk_id;
DROP INDEX IF EXISTS idx_disposisi_from_user_id;
DROP INDEX IF EXISTS idx_disposisi_to_user_id;

-- Drop foreign key
ALTER TABLE disposisi DROP CONSTRAINT IF EXISTS fk_disposisi_parent;

-- Drop columns
ALTER TABLE disposisi DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE disposisi DROP COLUMN IF EXISTS complete_at;
ALTER TABLE disposisi DROP COLUMN IF EXISTS baca_at;
ALTER TABLE disposisi DROP COLUMN IF EXISTS level;
ALTER TABLE disposisi DROP COLUMN IF EXISTS parent_disposisi_id;

COMMIT;


-- Migration: 000002_create_notifikasi_table.down.sql
BEGIN;

DROP TABLE IF EXISTS notifikasi;

COMMIT;


-- Migration: 000003_seed_jabatan_data.down.sql
BEGIN;

DELETE FROM jabatan WHERE nama_jabatan IN (
    'KEPALA_SEKOLAH', 'WAKIL_KEPALA', 'WAKIL_KURIKULUM', 'WAKIL_KESISWAAN',
    'WAKIL_SARANA', 'TATA_USAHA', 'GURU', 'STAFF', 'KEPALA_PERPUS', 'KEPALA_LAB', 'BK'
);

COMMIT;


-- ============================================
-- Verification Queries
-- ============================================

-- Check disposisi table structure
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name = 'disposisi' ORDER BY ordinal_position;

-- Check indexes
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'disposisi';

-- Check notifikasi table
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name = 'notifikasi' ORDER BY ordinal_position;

-- Check jabatan data
-- SELECT id_jabatan, nama_jabatan, level_akses FROM jabatan ORDER BY id_jabatan;


-- ============================================
-- Example Data for Testing
-- ============================================
-- Uncomment untuk seed test data

/*
BEGIN;

-- Assume sudah ada users dengan ID 1-10
-- User 1: TU
-- User 2: Kepala Sekolah
-- User 3: WAKA Kurikulum
-- User 4-5: Guru
-- User 6: WAKA Kesiswaan

-- 1. Create surat masuk
INSERT INTO surat_masuk (no_surat, perihal_surat, asal_surat, tanggal_surat, status_verifikasi)
VALUES ('001/2024', 'Rapat Kurikulum Semester 1', 'Dinas Pendidikan', NOW(), 'verified')
RETURNING id_surat_masuk;
-- Result: surat_id = 1

-- 2. Create root disposisi (TU -> Kepala Sekolah)
INSERT INTO disposisi (surat_masuk_id, from_user_id, to_user_id, parent_disposisi_id, level, status, catatan, dibaca)
VALUES (1, 1, 2, NULL, 0, 'pending', 'Silakan ditinjau dan diarahkan ke unit terkait', FALSE)
RETURNING id;
-- Result: disposisi_id = 1

-- 3. Kepala Sekolah forward ke WAKA Kurikulum (create child)
INSERT INTO disposisi (surat_masuk_id, from_user_id, to_user_id, parent_disposisi_id, level, status, catatan, dibaca)
VALUES (1, 2, 3, 1, 1, 'pending', 'Silakan handle rapat kurikulum', FALSE)
RETURNING id;
-- Result: disposisi_id = 2

-- 4. WAKA forward ke Guru (create child)
INSERT INTO disposisi (surat_masuk_id, from_user_id, to_user_id, parent_disposisi_id, level, status, catatan, dibaca)
VALUES (1, 3, 4, 2, 2, 'pending', 'Siapkan materi rapat', FALSE)
RETURNING id;
-- Result: disposisi_id = 3

-- 5. Update parent status
UPDATE disposisi SET status = 'forwarded' WHERE id = 1;
UPDATE disposisi SET status = 'forwarded' WHERE id = 2;

-- Test queries:
-- Get inbox untuk user 2 (Kepala Sekolah)
-- SELECT * FROM disposisi WHERE to_user_id = 2 AND deleted_at IS NULL ORDER BY created_at DESC;

-- Get sent items untuk user 1 (TU)
-- SELECT * FROM disposisi WHERE from_user_id = 1 AND deleted_at IS NULL ORDER BY created_at DESC;

-- Get full tree untuk surat 1
-- WITH RECURSIVE tree AS (
--   SELECT id, parent_disposisi_id, level, from_user_id, to_user_id, status FROM disposisi
--   WHERE id = 1 AND deleted_at IS NULL
--   UNION ALL
--   SELECT d.id, d.parent_disposisi_id, d.level, d.from_user_id, d.to_user_id, d.status
--   FROM disposisi d
--   INNER JOIN tree t ON d.parent_disposisi_id = t.id
--   WHERE d.deleted_at IS NULL
-- )
-- SELECT * FROM tree ORDER BY level, id;

COMMIT;
*/

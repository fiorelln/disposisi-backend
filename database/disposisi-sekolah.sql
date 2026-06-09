--
-- PostgreSQL database dump
--

\restrict yr7VxAr5WW7hRAyhaWHUcw6fufPo78uz3zT2B1cr64NVoOwnzzuPOcrD1l6uA2X

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: log_surat_masuk_changes(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.log_surat_masuk_changes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   IF OLD.status_verifikasi IS DISTINCT FROM NEW.status_verifikasi THEN
      INSERT INTO log (id_user, aksi, tabel_terkait, kolom_terkait, id_data, values_old, values_new)
      VALUES (
         NEW.user_verifikasi,
         'UPDATE status_verifikasi',
         'surat_masuk',
         'status_verifikasi',
         NEW.id_surat_masuk,
         OLD.status_verifikasi,
         NEW.status_verifikasi
      );
   END IF;
   RETURN NEW;
END;
$$;


--
-- Name: set_dibaca_disposisi_penerima(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_dibaca_disposisi_penerima() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   IF NEW.status = 'dibaca' AND OLD.status != 'dibaca' THEN
      NEW.read_at = CURRENT_TIMESTAMP;
   END IF;
   RETURN NEW;
END;
$$;


--
-- Name: set_tanggal_dibaca(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_tanggal_dibaca() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   IF NEW.status = 'dibaca' AND OLD.status != 'dibaca' THEN
      NEW.tanggal_dibaca = CURRENT_TIMESTAMP;
   END IF;
   RETURN NEW;
END;
$$;


--
-- Name: set_waktu_baca_notif(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.set_waktu_baca_notif() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   IF NEW.is_read = true AND OLD.is_read = false THEN
      NEW.waktu_baca = CURRENT_TIMESTAMP;
   END IF;
   RETURN NEW;
END;
$$;


--
-- Name: update_disposisi_aktif(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_disposisi_aktif() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   UPDATE surat_masuk 
   SET id_disposisi_aktif = NEW.id_disposisi
   WHERE id_surat_masuk = NEW.id_surat_masuk;
   
   RETURN NEW;
END;
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$BEGIN    NEW.updated_at = CURRENT_TIMESTAMP;    RETURN NEW;END;$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: disposisi; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.disposisi (
    id_disposisi integer NOT NULL,
    tanggapan_saran text,
    proses_lanjut text,
    koordinasi_konfirmasi text,
    id_surat_masuk integer,
    id_kepsek integer,
    id_penerima integer,
    tanggal_disposisi timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    status_disposisi character varying(50) DEFAULT 'belum_dibaca'::character varying,
    status_approval character varying(50) DEFAULT 'menunggu'::character varying,
    approval_at timestamp without time zone,
    catatan_kepsek text,
    id_jabatan_penerima integer,
    CONSTRAINT chk_status_approval CHECK (((status_approval)::text = ANY ((ARRAY['menunggu'::character varying, 'disetujui'::character varying, 'ditolak'::character varying])::text[]))),
    CONSTRAINT disposisi_status_disposisi_check CHECK (((status_disposisi)::text = ANY ((ARRAY['belum_dibaca'::character varying, 'dibaca'::character varying, 'sedang_dikerjakan'::character varying, 'selesai'::character varying])::text[])))
);


--
-- Name: disposisi_id_disposisi_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.disposisi_id_disposisi_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: disposisi_id_disposisi_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.disposisi_id_disposisi_seq OWNED BY public.disposisi.id_disposisi;


--
-- Name: distribusi_sm; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.distribusi_sm (
    id_penerima_disposisi integer CONSTRAINT disposisi_penerima_id_penerima_disposisi_not_null NOT NULL,
    id_disposisi integer CONSTRAINT disposisi_penerima_id_disposisi_not_null NOT NULL,
    id_user integer,
    id_jabatan integer,
    read_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    status character varying(50) DEFAULT 'belum_dibaca'::character varying,
    id_waka integer,
    id_distribusi_parent integer,
    CONSTRAINT chk_xor_disposisi_penerima CHECK ((((id_user IS NOT NULL) AND (id_jabatan IS NULL)) OR ((id_user IS NULL) AND (id_jabatan IS NOT NULL)))),
    CONSTRAINT disposisi_penerima_status_check CHECK (((status)::text = ANY (ARRAY[('belum_dibaca'::character varying)::text, ('dibaca'::character varying)::text, ('diteruskan_waka'::character varying)::text, ('selesai'::character varying)::text])))
);


--
-- Name: disposisi_penerima_id_penerima_disposisi_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.disposisi_penerima_id_penerima_disposisi_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: disposisi_penerima_id_penerima_disposisi_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.disposisi_penerima_id_penerima_disposisi_seq OWNED BY public.distribusi_sm.id_penerima_disposisi;


--
-- Name: distribusi_sk; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.distribusi_sk (
    id_distribusi integer CONSTRAINT distribusi_surat_keluar_id_distribusi_not_null NOT NULL,
    id_sk integer CONSTRAINT distribusi_surat_keluar_id_surat_keluar_not_null NOT NULL,
    id_user integer,
    status character varying(50) DEFAULT 'belum_dibaca'::character varying,
    distribute_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    read_at timestamp without time zone,
    catatan text,
    id_jabatan integer,
    CONSTRAINT chk_xor_distribusi_penerima CHECK ((((id_user IS NOT NULL) AND (id_jabatan IS NULL)) OR ((id_user IS NULL) AND (id_jabatan IS NOT NULL)))),
    CONSTRAINT distribusi_surat_keluar_status_check CHECK (((status)::text = ANY ((ARRAY['belum_dibaca'::character varying, 'dibaca'::character varying, 'selesai'::character varying])::text[])))
);


--
-- Name: distribusi_surat_keluar_id_distribusi_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.distribusi_surat_keluar_id_distribusi_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: distribusi_surat_keluar_id_distribusi_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.distribusi_surat_keluar_id_distribusi_seq OWNED BY public.distribusi_sk.id_distribusi;


--
-- Name: jabatan; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jabatan (
    id_jabatan integer NOT NULL,
    nama_jabatan character varying(50) NOT NULL,
    level_akses character varying(20),
    CONSTRAINT jabatan_level_akses_check CHECK (((level_akses)::text = ANY (ARRAY[('kepsek'::character varying)::text, ('admin'::character varying)::text, ('pegawai'::character varying)::text, ('waka'::character varying)::text, ('user'::character varying)::text])))
);


--
-- Name: jabatan_id_jabatan_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jabatan_id_jabatan_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jabatan_id_jabatan_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jabatan_id_jabatan_seq OWNED BY public.jabatan.id_jabatan;


--
-- Name: log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.log (
    id_log integer CONSTRAINT log_aktivitas_id_log_not_null NOT NULL,
    id_user integer,
    aksi character varying(200),
    tabel_terkait character varying(100),
    kolom_terkait character varying(100),
    id_data integer,
    values_old text,
    values_new text,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: log_aktivitas_id_log_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.log_aktivitas_id_log_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: log_aktivitas_id_log_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.log_aktivitas_id_log_seq OWNED BY public.log.id_log;


--
-- Name: log_distribusi; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.log_distribusi (
    id_riwayat integer CONSTRAINT riwayat_alur_surat_id_riwayat_not_null NOT NULL,
    id_sm integer,
    id_sk integer,
    status_asal character varying(50),
    status_tujuan character varying(50),
    id_user integer,
    catatan text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_xor_surat CHECK ((((id_sm IS NOT NULL) AND (id_sk IS NULL)) OR ((id_sk IS NULL) AND (id_sk IS NOT NULL))))
);


--
-- Name: notifikasi; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifikasi (
    id_notifikasi integer NOT NULL,
    id_penerima integer NOT NULL,
    id_pengirim integer,
    jenis character varying(30),
    judul character varying(300) NOT NULL,
    pesan text,
    is_read boolean DEFAULT false,
    waktu_baca timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    link_url character varying(500),
    tipe_referensi character varying(20),
    CONSTRAINT notifikasi_jenis_check CHECK (((jenis)::text = ANY ((ARRAY['surat_masuk_baru'::character varying, 'surat_keluar_baru'::character varying, 'surat_disetujui'::character varying, 'surat_ditolak'::character varying, 'surat_masuk_dikonfirmasi'::character varying, 'surat_keluar_dikonfirmasi'::character varying, 'permintaan_persetujuan_akun'::character varying])::text[])))
);


--
-- Name: notifikasi_id_notifikasi_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifikasi_id_notifikasi_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifikasi_id_notifikasi_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifikasi_id_notifikasi_seq OWNED BY public.notifikasi.id_notifikasi;


--
-- Name: otp; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.otp (
    id_otp integer NOT NULL,
    id_user integer NOT NULL,
    kode_otp character varying(10) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    is_used boolean DEFAULT false
);


--
-- Name: otp_id_otp_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.otp_id_otp_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: otp_id_otp_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.otp_id_otp_seq OWNED BY public.otp.id_otp;


--
-- Name: riwayat_alur_surat_id_riwayat_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.riwayat_alur_surat_id_riwayat_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: riwayat_alur_surat_id_riwayat_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.riwayat_alur_surat_id_riwayat_seq OWNED BY public.log_distribusi.id_riwayat;


--
-- Name: surat_keluar; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.surat_keluar (
    id_surat_keluar integer NOT NULL,
    kode_surat integer NOT NULL,
    no_surat character varying(100) NOT NULL,
    perihal character varying(200) NOT NULL,
    catatan character varying(300),
    tanggal_surat date NOT NULL,
    file_pdf character varying(500),
    status_verifikasi character varying(50) DEFAULT 'menunggu'::character varying,
    user_verifikasi integer,
    tanggal_verifikasi timestamp without time zone,
    tujuan character varying(200),
    catatan_verifikasi text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    status_alur character varying(50) DEFAULT 'diterima_tu'::character varying,
    CONSTRAINT surat_keluar_status_check CHECK (((status_verifikasi)::text = ANY ((ARRAY['menunggu'::character varying, 'disetujui'::character varying, 'ditolak'::character varying])::text[])))
);


--
-- Name: surat_keluar_id_surat_keluar_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.surat_keluar_id_surat_keluar_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: surat_keluar_id_surat_keluar_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.surat_keluar_id_surat_keluar_seq OWNED BY public.surat_keluar.id_surat_keluar;


--
-- Name: surat_masuk; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.surat_masuk (
    id_surat_masuk integer NOT NULL,
    no_surat character varying(100) NOT NULL,
    perihal_surat character varying(200) NOT NULL,
    asal_surat character varying(200) NOT NULL,
    tanggal_surat date NOT NULL,
    file_pdf character varying(500),
    tanggal_diterima date DEFAULT CURRENT_DATE,
    status_verifikasi character varying(50) DEFAULT 'menunggu'::character varying,
    user_verifikasi integer,
    tanggal_verifikasi timestamp without time zone,
    catatan_verifikasi text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    id_disposisi_aktif integer,
    status_alur character varying(50) DEFAULT 'diterima_tu'::character varying,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_status_alur_sm CHECK (((status_alur)::text = ANY (ARRAY[('diterima_tu'::character varying)::text, ('disposisi_kepsek'::character varying)::text, ('didistribusikan_waka'::character varying)::text, ('selesai'::character varying)::text]))),
    CONSTRAINT surat_masuk_status_check CHECK (((status_verifikasi)::text = ANY ((ARRAY['menunggu'::character varying, 'disetujui'::character varying, 'ditolak'::character varying])::text[])))
);


--
-- Name: surat_masuk_id_surat_masuk_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.surat_masuk_id_surat_masuk_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: surat_masuk_id_surat_masuk_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.surat_masuk_id_surat_masuk_seq OWNED BY public.surat_masuk.id_surat_masuk;


--
-- Name: user_jabatan; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_jabatan (
    id_user integer NOT NULL,
    id_jabatan integer NOT NULL,
    is_primary boolean DEFAULT false
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id_user integer NOT NULL,
    nama character varying(100) NOT NULL,
    email character varying(100) NOT NULL,
    password character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: users_id_user_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_user_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_user_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_user_seq OWNED BY public.users.id_user;


--
-- Name: disposisi id_disposisi; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi ALTER COLUMN id_disposisi SET DEFAULT nextval('public.disposisi_id_disposisi_seq'::regclass);


--
-- Name: distribusi_sk id_distribusi; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sk ALTER COLUMN id_distribusi SET DEFAULT nextval('public.distribusi_surat_keluar_id_distribusi_seq'::regclass);


--
-- Name: distribusi_sm id_penerima_disposisi; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm ALTER COLUMN id_penerima_disposisi SET DEFAULT nextval('public.disposisi_penerima_id_penerima_disposisi_seq'::regclass);


--
-- Name: jabatan id_jabatan; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jabatan ALTER COLUMN id_jabatan SET DEFAULT nextval('public.jabatan_id_jabatan_seq'::regclass);


--
-- Name: log id_log; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log ALTER COLUMN id_log SET DEFAULT nextval('public.log_aktivitas_id_log_seq'::regclass);


--
-- Name: log_distribusi id_riwayat; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log_distribusi ALTER COLUMN id_riwayat SET DEFAULT nextval('public.riwayat_alur_surat_id_riwayat_seq'::regclass);


--
-- Name: notifikasi id_notifikasi; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifikasi ALTER COLUMN id_notifikasi SET DEFAULT nextval('public.notifikasi_id_notifikasi_seq'::regclass);


--
-- Name: otp id_otp; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp ALTER COLUMN id_otp SET DEFAULT nextval('public.otp_id_otp_seq'::regclass);


--
-- Name: surat_keluar id_surat_keluar; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_keluar ALTER COLUMN id_surat_keluar SET DEFAULT nextval('public.surat_keluar_id_surat_keluar_seq'::regclass);


--
-- Name: surat_masuk id_surat_masuk; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_masuk ALTER COLUMN id_surat_masuk SET DEFAULT nextval('public.surat_masuk_id_surat_masuk_seq'::regclass);


--
-- Name: users id_user; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id_user SET DEFAULT nextval('public.users_id_user_seq'::regclass);


--
-- Data for Name: disposisi; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.disposisi (id_disposisi, tanggapan_saran, proses_lanjut, koordinasi_konfirmasi, id_surat_masuk, id_kepsek, id_penerima, tanggal_disposisi, status_disposisi, status_approval, approval_at, catatan_kepsek, id_jabatan_penerima) FROM stdin;
\.


--
-- Data for Name: distribusi_sk; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.distribusi_sk (id_distribusi, id_sk, id_user, status, distribute_at, read_at, catatan, id_jabatan) FROM stdin;
\.


--
-- Data for Name: distribusi_sm; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.distribusi_sm (id_penerima_disposisi, id_disposisi, id_user, id_jabatan, read_at, created_at, status, id_waka, id_distribusi_parent) FROM stdin;
\.


--
-- Data for Name: jabatan; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.jabatan (id_jabatan, nama_jabatan, level_akses) FROM stdin;
1	kepala sekolah	kepsek
2	admin	admin
3	pegawai	pegawai
9	bkk	user
11	kapro rpl	user
12	kapro tkj	user
13	kapro dkv	user
15	kapro ei	user
16	kapro mt	user
17	kapro av	user
18	kapro bc	user
14	kapro an	user
19	bk	user
20	prakerin	user
22	koordinator bk	user
23	koordinator bkk	user
5	waka kesiswaan	waka
6	waka kurikulum	waka
7	waka sarpras	waka
8	waka humas	waka
21	koordinator waka	waka
\.


--
-- Data for Name: log; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.log (id_log, id_user, aksi, tabel_terkait, kolom_terkait, id_data, values_old, values_new, updated_at) FROM stdin;
\.


--
-- Data for Name: log_distribusi; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.log_distribusi (id_riwayat, id_sm, id_sk, status_asal, status_tujuan, id_user, catatan, created_at) FROM stdin;
\.


--
-- Data for Name: notifikasi; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.notifikasi (id_notifikasi, id_penerima, id_pengirim, jenis, judul, pesan, is_read, waktu_baca, created_at, link_url, tipe_referensi) FROM stdin;
\.


--
-- Data for Name: otp; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.otp (id_otp, id_user, kode_otp, expires_at, created_at, is_used) FROM stdin;
\.


--
-- Data for Name: surat_keluar; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.surat_keluar (id_surat_keluar, kode_surat, no_surat, perihal, catatan, tanggal_surat, file_pdf, status_verifikasi, user_verifikasi, tanggal_verifikasi, tujuan, catatan_verifikasi, created_at, updated_at, status_alur) FROM stdin;
\.


--
-- Data for Name: surat_masuk; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.surat_masuk (id_surat_masuk, no_surat, perihal_surat, asal_surat, tanggal_surat, file_pdf, tanggal_diterima, status_verifikasi, user_verifikasi, tanggal_verifikasi, catatan_verifikasi, created_at, id_disposisi_aktif, status_alur, updated_at) FROM stdin;
\.


--
-- Data for Name: user_jabatan; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_jabatan (id_user, id_jabatan, is_primary) FROM stdin;
1	1	t
2	2	t
3	3	t
4	11	t
5	2	t
6	9	t
6	8	f
7	8	t
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id_user, nama, email, password, created_at) FROM stdin;
1	dummy_kepsek	dummy_kepsek@gmail.com	12345	2026-05-25 19:44:51.727343
2	dummy_admin	dummy_admin@gmail.com	12345	2026-05-25 19:46:19.919197
3	dummy_pegawai	dummy_peagwai@gmail.com	12345	2026-05-25 19:46:19.919197
4	dummy_rpl	dummy_rpl@gmail.com	12345	2026-05-25 19:46:19.919197
5	dummy_admin2	dummy_admin2@gmail.com	12345	2026-05-25 19:47:04.095179
6	dummy_bkk	dummy_bkk@gmail.com	12345	2026-05-25 19:51:32.59051
7	dummy_wakahumas	dummy_wakahumas@gmail.com	12345	2026-06-04 20:16:12.231802
\.


--
-- Name: disposisi_id_disposisi_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.disposisi_id_disposisi_seq', 1, false);


--
-- Name: disposisi_penerima_id_penerima_disposisi_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.disposisi_penerima_id_penerima_disposisi_seq', 1, false);


--
-- Name: distribusi_surat_keluar_id_distribusi_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.distribusi_surat_keluar_id_distribusi_seq', 1, false);


--
-- Name: jabatan_id_jabatan_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.jabatan_id_jabatan_seq', 1, true);


--
-- Name: log_aktivitas_id_log_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.log_aktivitas_id_log_seq', 1, false);


--
-- Name: notifikasi_id_notifikasi_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.notifikasi_id_notifikasi_seq', 1, false);


--
-- Name: otp_id_otp_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.otp_id_otp_seq', 1, false);


--
-- Name: riwayat_alur_surat_id_riwayat_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.riwayat_alur_surat_id_riwayat_seq', 1, false);


--
-- Name: surat_keluar_id_surat_keluar_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.surat_keluar_id_surat_keluar_seq', 1, false);


--
-- Name: surat_masuk_id_surat_masuk_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.surat_masuk_id_surat_masuk_seq', 1, false);


--
-- Name: users_id_user_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_id_user_seq', 6, true);


--
-- Name: distribusi_sm disposisi_penerima_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT disposisi_penerima_pkey PRIMARY KEY (id_penerima_disposisi);


--
-- Name: disposisi disposisi_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi
    ADD CONSTRAINT disposisi_pkey PRIMARY KEY (id_disposisi);


--
-- Name: distribusi_sk distribusi_surat_keluar_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sk
    ADD CONSTRAINT distribusi_surat_keluar_pkey PRIMARY KEY (id_distribusi);


--
-- Name: jabatan jabatan_nama_jabatan_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jabatan
    ADD CONSTRAINT jabatan_nama_jabatan_key UNIQUE (nama_jabatan);


--
-- Name: jabatan jabatan_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jabatan
    ADD CONSTRAINT jabatan_pkey PRIMARY KEY (id_jabatan);


--
-- Name: log log_aktivitas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_aktivitas_pkey PRIMARY KEY (id_log);


--
-- Name: notifikasi notifikasi_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifikasi
    ADD CONSTRAINT notifikasi_pkey PRIMARY KEY (id_notifikasi);


--
-- Name: otp otp_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp
    ADD CONSTRAINT otp_pkey PRIMARY KEY (id_otp);


--
-- Name: log_distribusi riwayat_alur_surat_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log_distribusi
    ADD CONSTRAINT riwayat_alur_surat_pkey PRIMARY KEY (id_riwayat);


--
-- Name: surat_keluar surat_keluar_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_keluar
    ADD CONSTRAINT surat_keluar_pkey PRIMARY KEY (id_surat_keluar);


--
-- Name: surat_masuk surat_masuk_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_masuk
    ADD CONSTRAINT surat_masuk_pkey PRIMARY KEY (id_surat_masuk);


--
-- Name: distribusi_sm uq_disposisi_penerima; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT uq_disposisi_penerima UNIQUE (id_disposisi, id_user, id_jabatan);


--
-- Name: user_jabatan user_jabatan_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_jabatan
    ADD CONSTRAINT user_jabatan_pkey PRIMARY KEY (id_user, id_jabatan);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id_user);


--
-- Name: disposisi trg_disposisi_baru; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_disposisi_baru AFTER INSERT ON public.disposisi FOR EACH ROW EXECUTE FUNCTION public.update_disposisi_aktif();


--
-- Name: distribusi_sk trg_distribusi_dibaca; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_distribusi_dibaca BEFORE UPDATE ON public.distribusi_sk FOR EACH ROW EXECUTE FUNCTION public.set_tanggal_dibaca();


--
-- Name: distribusi_sk trg_distribusi_sk_dibaca; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_distribusi_sk_dibaca BEFORE UPDATE ON public.distribusi_sk FOR EACH ROW EXECUTE FUNCTION public.set_tanggal_dibaca();


--
-- Name: surat_masuk trg_log_surat_masuk; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_log_surat_masuk AFTER UPDATE ON public.surat_masuk FOR EACH ROW EXECUTE FUNCTION public.log_surat_masuk_changes();


--
-- Name: notifikasi trg_notifikasi_baca; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_notifikasi_baca BEFORE UPDATE ON public.notifikasi FOR EACH ROW EXECUTE FUNCTION public.set_waktu_baca_notif();


--
-- Name: surat_keluar trg_surat_keluar_updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_surat_keluar_updated BEFORE UPDATE ON public.surat_keluar FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: surat_masuk trg_surat_masuk_updated; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_surat_masuk_updated BEFORE UPDATE ON public.surat_masuk FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: disposisi disposisi_id_jabatan_penerima_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi
    ADD CONSTRAINT disposisi_id_jabatan_penerima_fkey FOREIGN KEY (id_jabatan_penerima) REFERENCES public.jabatan(id_jabatan);


--
-- Name: distribusi_sm disposisi_penerima_disposisi_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT disposisi_penerima_disposisi_fkey FOREIGN KEY (id_disposisi) REFERENCES public.disposisi(id_disposisi) ON DELETE CASCADE;


--
-- Name: distribusi_sm disposisi_penerima_jabatan_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT disposisi_penerima_jabatan_fkey FOREIGN KEY (id_jabatan) REFERENCES public.jabatan(id_jabatan) ON DELETE CASCADE;


--
-- Name: distribusi_sm disposisi_penerima_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT disposisi_penerima_user_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user) ON DELETE CASCADE;


--
-- Name: distribusi_sk distribusi_sk_id_jabatan_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sk
    ADD CONSTRAINT distribusi_sk_id_jabatan_fkey FOREIGN KEY (id_jabatan) REFERENCES public.jabatan(id_jabatan);


--
-- Name: distribusi_sm distribusi_sm_id_waka_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT distribusi_sm_id_waka_fkey FOREIGN KEY (id_waka) REFERENCES public.users(id_user);


--
-- Name: distribusi_sk distribusi_surat_keluar_id_penerima_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sk
    ADD CONSTRAINT distribusi_surat_keluar_id_penerima_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user);


--
-- Name: distribusi_sk distribusi_surat_keluar_id_surat_keluar_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sk
    ADD CONSTRAINT distribusi_surat_keluar_id_surat_keluar_fkey FOREIGN KEY (id_sk) REFERENCES public.surat_keluar(id_surat_keluar) ON DELETE CASCADE;


--
-- Name: disposisi fk_disposisi_kepsek; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi
    ADD CONSTRAINT fk_disposisi_kepsek FOREIGN KEY (id_kepsek) REFERENCES public.users(id_user);


--
-- Name: disposisi fk_disposisi_penerima; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi
    ADD CONSTRAINT fk_disposisi_penerima FOREIGN KEY (id_penerima) REFERENCES public.users(id_user);


--
-- Name: disposisi fk_disposisi_surat_masuk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.disposisi
    ADD CONSTRAINT fk_disposisi_surat_masuk FOREIGN KEY (id_surat_masuk) REFERENCES public.surat_masuk(id_surat_masuk);


--
-- Name: distribusi_sm fk_distribusi_parent; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.distribusi_sm
    ADD CONSTRAINT fk_distribusi_parent FOREIGN KEY (id_distribusi_parent) REFERENCES public.distribusi_sm(id_penerima_disposisi) ON DELETE SET NULL;


--
-- Name: log_distribusi fk_riwayat_sk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log_distribusi
    ADD CONSTRAINT fk_riwayat_sk FOREIGN KEY (id_sk) REFERENCES public.surat_keluar(id_surat_keluar) ON DELETE CASCADE;


--
-- Name: log_distribusi fk_riwayat_sm; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log_distribusi
    ADD CONSTRAINT fk_riwayat_sm FOREIGN KEY (id_sm) REFERENCES public.surat_masuk(id_surat_masuk) ON DELETE CASCADE;


--
-- Name: surat_masuk fk_surat_masuk_disposisi_aktif; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_masuk
    ADD CONSTRAINT fk_surat_masuk_disposisi_aktif FOREIGN KEY (id_disposisi_aktif) REFERENCES public.disposisi(id_disposisi) ON DELETE SET NULL;


--
-- Name: surat_masuk fk_surat_masuk_verifikasi; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_masuk
    ADD CONSTRAINT fk_surat_masuk_verifikasi FOREIGN KEY (user_verifikasi) REFERENCES public.users(id_user);


--
-- Name: log log_aktivitas_id_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log
    ADD CONSTRAINT log_aktivitas_id_user_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user);


--
-- Name: notifikasi notifikasi_id_penerima_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifikasi
    ADD CONSTRAINT notifikasi_id_penerima_fkey FOREIGN KEY (id_penerima) REFERENCES public.users(id_user);


--
-- Name: notifikasi notifikasi_id_pengirim_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifikasi
    ADD CONSTRAINT notifikasi_id_pengirim_fkey FOREIGN KEY (id_pengirim) REFERENCES public.users(id_user);


--
-- Name: otp otp_id_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.otp
    ADD CONSTRAINT otp_id_user_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user);


--
-- Name: log_distribusi riwayat_alur_surat_id_user_pelaku_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.log_distribusi
    ADD CONSTRAINT riwayat_alur_surat_id_user_pelaku_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user);


--
-- Name: surat_keluar surat_keluar_verifikasi_oleh_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.surat_keluar
    ADD CONSTRAINT surat_keluar_verifikasi_oleh_fkey FOREIGN KEY (user_verifikasi) REFERENCES public.users(id_user);


--
-- Name: user_jabatan user_jabatan_id_jabatan_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_jabatan
    ADD CONSTRAINT user_jabatan_id_jabatan_fkey FOREIGN KEY (id_jabatan) REFERENCES public.jabatan(id_jabatan);


--
-- Name: user_jabatan user_jabatan_id_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_jabatan
    ADD CONSTRAINT user_jabatan_id_user_fkey FOREIGN KEY (id_user) REFERENCES public.users(id_user);


--
-- PostgreSQL database dump complete
--

\unrestrict yr7VxAr5WW7hRAyhaWHUcw6fufPo78uz3zT2B1cr64NVoOwnzzuPOcrD1l6uA2X


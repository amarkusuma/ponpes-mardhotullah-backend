package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/config"
	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/penilaian"
)

// Seed mengisi data awal. Aman dijalankan berulang (idempotent via FirstOrCreate).
//
//   - Semua environment : permission & role (dibutuhkan sistem RBAC).
//   - production        : admin awal dari ADMIN_EMAIL/ADMIN_PASSWORD, hanya bila
//     belum ada user ber-role admin. Tidak ada data dummy.
//   - development       : seluruh data dummy (admin, guru, santri, wali, kelas, dst).
func Seed(db *gorm.DB, cfg *config.Config) error {
	log.Printf("▶ Seeding dimulai (APP_ENV=%s)...", cfg.AppEnv)

	// ── 1. Permissions ────────────────────────────────────────────
	permissions := []string{
		"user:read", "user:write", "user:delete",
		"santri:read", "santri:write", "santri:delete",
		"keuangan:read", "keuangan:write",
		"akademik:read", "akademik:write",
		"konten:read", "konten:write",
	}
	permMap := make(map[string]model.Permission)
	for _, name := range permissions {
		p := model.Permission{}
		db.FirstOrCreate(&p, model.Permission{Name: name})
		permMap[name] = p
	}

	// ── 2. Roles ──────────────────────────────────────────────────
	roleAdmin := seedRole(db, "admin", allPermissions(permMap))
	roleGuru := seedRole(db, "guru", permissionsFor(permMap,
		"santri:read", "akademik:read", "akademik:write", "keuangan:read",
	))
	roleWali := seedRole(db, "wali", permissionsFor(permMap,
		"santri:read", "keuangan:read",
	))
	roleSantri := seedRole(db, "santri", permissionsFor(permMap,
		"santri:read", "akademik:read",
	))
	if cfg.IsProduction() {
		if err := seedProductionAdmin(db, cfg, roleAdmin); err != nil {
			return err
		}
		log.Println("✅ Seeding production selesai (tanpa data dummy).")
		return nil
	}

	// ── 3. Tahun Ajaran ──────────────────────────────────────────
	ta := akademik.TahunAjaran{}
	db.FirstOrCreate(&ta, akademik.TahunAjaran{TahunMulai: 2024, TahunSelesai: 2025, IsActive: true})

	// ── 4. Mata Pelajaran ─────────────────────────────────────────
	mapelList := []akademik.MataPelajaran{
		{Nama: "Al-Quran & Tajwid", Kode: "AQT", Deskripsi: "Hafalan dan tilawah Al-Quran"},
		{Nama: "Fiqih", Kode: "FQH", Deskripsi: "Hukum Islam praktis"},
		{Nama: "Bahasa Arab", Kode: "BAR", Deskripsi: "Nahwu, Sharaf, dan percakapan"},
		{Nama: "Matematika", Kode: "MTK", Deskripsi: "Matematika umum"},
		{Nama: "Bahasa Indonesia", Kode: "BIN", Deskripsi: "Bahasa Indonesia"},
	}
	for i := range mapelList {
		db.FirstOrCreate(&mapelList[i], akademik.MataPelajaran{Kode: mapelList[i].Kode})
	}

	// ── 5. Jenis Penilaian ────────────────────────────────────────
	jenisList := []string{"UTS", "UAS", "Harian", "Praktik", "Tugas"}
	jenisMap := make(map[string]penilaian.JenisPenilaian)
	for _, nama := range jenisList {
		jp := penilaian.JenisPenilaian{}
		db.FirstOrCreate(&jp, penilaian.JenisPenilaian{Nama: nama})
		jenisMap[nama] = jp
	}

	// ── 6. Admin ──────────────────────────────────────────────────
	adminUser := seedUser(db, "Administrator", "admin@ponpes.id", "Admin123!", roleAdmin)

	// ── 7. Guru (3 orang) ─────────────────────────────────────────
	guruData := []struct {
		nama, email, nik, spesialisasi string
	}{
		{"Ust. Ahmad Fauzi", "ahmad.fauzi@ponpes.id", "3201011001001001", "Al-Quran & Tajwid"},
		{"Ust. Hasan Basri", "hasan.basri@ponpes.id", "3201011002001002", "Fiqih & Ushul Fiqih"},
		{"Ust. Yusuf Ali", "yusuf.ali@ponpes.id", "3201011003001003", "Bahasa Arab"},
	}
	guruList := make([]akademik.Guru, 0, len(guruData))
	for _, g := range guruData {
		u := seedUser(db, g.nama, g.email, "Guru123!", roleGuru)
		guru := akademik.Guru{}
		db.FirstOrCreate(&guru, akademik.Guru{UserID: u.ID})
		if guru.NIK == "" {
			guru.NIK = g.nik
			guru.Spesialisasi = g.spesialisasi
			db.Save(&guru)
		}
		guruList = append(guruList, guru)
	}

	// ── 8. Santri (10 orang) ──────────────────────────────────────
	santriData := []struct {
		nama, email, nik, tempat string
		tgl                      time.Time
	}{
		{"Muhammad Rizki", "rizki@santri.ponpes.id", "3201010301100001", "Bandung", dob(2010, 3, 15)},
		{"Abdullah Fatih", "fatih@santri.ponpes.id", "3171010307100002", "Jakarta", dob(2010, 7, 22)},
		{"Ahmad Zaki", "zaki@santri.ponpes.id", "3578010501110003", "Surabaya", dob(2011, 1, 5)},
		{"Umar Faruq", "faruq@santri.ponpes.id", "3374013011100004", "Semarang", dob(2010, 11, 30)},
		{"Hasan Albana", "hasan@santri.ponpes.id", "3471011804110005", "Yogyakarta", dob(2011, 4, 18)},
		{"Ibrahim Khalil", "ibrahim@santri.ponpes.id", "3573011209100006", "Malang", dob(2010, 9, 12)},
		{"Ismail Hakim", "ismail@santri.ponpes.id", "1271012506110007", "Medan", dob(2011, 6, 25)},
		{"Yusuf Siddiq", "yusuf@santri.ponpes.id", "7371010812100008", "Makassar", dob(2010, 12, 8)},
		{"Musa Rasyid", "musa@santri.ponpes.id", "1371011402110009", "Padang", dob(2011, 2, 14)},
		{"Isa Ramadhan", "isa@santri.ponpes.id", "1671012008100010", "Palembang", dob(2010, 8, 20)},
	}
	santriList := make([]akademik.Santri, 0, len(santriData))
	for _, s := range santriData {
		u := seedUser(db, s.nama, s.email, "Santri123!", roleSantri)
		santri := akademik.Santri{}
		db.FirstOrCreate(&santri, akademik.Santri{UserID: u.ID})
		if santri.NIK == "" {
			santri.NIK = s.nik
			santri.TempatLahir = s.tempat
			santri.TanggalLahir = s.tgl
			santri.Alamat = fmt.Sprintf("Jl. Santri No. %d, %s", len(santriList)+1, s.tempat)
			santri.Status = akademik.StatusAktif
			db.Save(&santri)
		}
		santriList = append(santriList, santri)
	}

	// ── 9. Wali (3 orang) ─────────────────────────────────────────
	waliData := []struct {
		nama, email, nik, telepon, pekerjaan, hubungan, alamat string
		santriIdx                                               []int
	}{
		{"Bapak Rizki Sr.", "wali.rizki@gmail.com", "3201011505700001", "081234560001", "Wiraswasta", "Ayah", "Jl. Mawar No. 1, Bandung", []int{0, 1}},
		{"Bapak Ahmad Sr.", "wali.ahmad@gmail.com", "3578011206680002", "081234560002", "PNS", "Ayah", "Jl. Kenanga No. 5, Surabaya", []int{2, 3, 4}},
		{"Ibu Siti Rahma", "wali.siti@gmail.com", "3471016507720003", "081234560003", "Ibu Rumah Tangga", "Ibu", "Jl. Melati No. 10, Yogyakarta", []int{5, 6, 7, 8, 9}},
	}
	for _, w := range waliData {
		u := seedUser(db, w.nama, w.email, "Wali123!", roleWali)
		wali := akademik.Wali{}
		db.FirstOrCreate(&wali, akademik.Wali{UserID: u.ID})
		if wali.Telepon == "" {
			wali.NIK = w.nik
			wali.Telepon = w.telepon
			wali.Pekerjaan = w.pekerjaan
			wali.Hubungan = w.hubungan
			wali.Alamat = w.alamat
			db.Save(&wali)
		}
		// Relasi wali ↔ santri via many2many
		for _, idx := range w.santriIdx {
			db.Model(&wali).Association("Santri").Append(&santriList[idx])
		}
	}

	// ── 10. Kelas ─────────────────────────────────────────────────
	kelasData := []struct {
		nama    string
		tingkat int
		guruIdx int
	}{
		{"Kelas 1A", 1, 0},
		{"Kelas 2A", 2, 1},
		{"Kelas 3A", 3, 2},
	}
	kelasList := make([]akademik.Kelas, 0, len(kelasData))
	for i, k := range kelasData {
		kelas := akademik.Kelas{}
		db.FirstOrCreate(&kelas, akademik.Kelas{NamaKelas: k.nama, TahunAjaranID: ta.ID})
		if kelas.Tingkat == 0 {
			kelas.Tingkat = k.tingkat
			kelas.WaliKelasID = &guruList[k.guruIdx].ID
			db.Save(&kelas)
		}
		// Distribusi santri ke kelas (round-robin sederhana)
		for j, santri := range santriList {
			if j%len(kelasData) == i {
				ks := akademik.KelasSantri{}
				db.FirstOrCreate(&ks, akademik.KelasSantri{
					SantriID: santri.ID,
					KelasID:  kelas.ID,
				})
			}
		}
		kelasList = append(kelasList, kelas)
	}

	_ = adminUser
	log.Println("✅ Seeding selesai.")
	return nil
}

// ─────────────────────────────────────────────
//  Helpers
// ─────────────────────────────────────────────

// seedProductionAdmin membuat admin awal hanya jika belum ada user ber-role admin.
// Tidak pernah mengubah password user yang sudah ada.
func seedProductionAdmin(db *gorm.DB, cfg *config.Config, roleAdmin model.Role) error {
	var adminCount int64
	err := db.Table("user_roles").
		Joins("JOIN users ON users.id = user_roles.user_id AND users.deleted_at IS NULL").
		Where("user_roles.role_id = ?", roleAdmin.ID).
		Count(&adminCount).Error
	if err != nil {
		return fmt.Errorf("cek admin: %w", err)
	}
	if adminCount > 0 {
		return nil
	}

	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		log.Println("⚠ Belum ada user admin. Set ADMIN_EMAIL & ADMIN_PASSWORD untuk membuat admin awal.")
		return nil
	}
	if len(cfg.AdminPassword) < 8 {
		return errors.New("ADMIN_PASSWORD minimal 8 karakter")
	}

	var existing model.User
	if err := db.Where("email = ?", cfg.AdminEmail).First(&existing).Error; err == nil {
		return fmt.Errorf("email %s sudah dipakai user lain; tidak dijadikan admin otomatis", cfg.AdminEmail)
	}

	seedUser(db, cfg.AdminName, cfg.AdminEmail, cfg.AdminPassword, roleAdmin)
	log.Printf("✅ Admin awal dibuat: %s", cfg.AdminEmail)
	return nil
}

func seedUser(db *gorm.DB, name, email, password string, role model.Role) model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u := model.User{}
	db.FirstOrCreate(&u, model.User{Email: email})
	if u.Name == "" {
		u.Name = name
		u.Password = string(hash)
		u.IsActive = true
		db.Save(&u)
	}
	db.Model(&u).Association("Roles").Append(&role)
	return u
}

func seedRole(db *gorm.DB, name string, perms []model.Permission) model.Role {
	r := model.Role{}
	db.FirstOrCreate(&r, model.Role{Name: name})
	if len(perms) > 0 {
		db.Model(&r).Association("Permissions").Replace(perms)
	}
	return r
}

func allPermissions(m map[string]model.Permission) []model.Permission {
	perms := make([]model.Permission, 0, len(m))
	for _, p := range m {
		perms = append(perms, p)
	}
	return perms
}

func permissionsFor(m map[string]model.Permission, names ...string) []model.Permission {
	perms := make([]model.Permission, 0, len(names))
	for _, n := range names {
		if p, ok := m[n]; ok {
			perms = append(perms, p)
		}
	}
	return perms
}

func dob(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

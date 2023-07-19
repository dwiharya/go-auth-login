// main.go
package main

import (
	"auth-register/database"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var store = sessions.NewCookieStore([]byte("secret"))

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Panggil fungsi NewDBConnection untuk menghubungkan ke database
	_, err = database.NewDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/register", registerHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/logout", logoutHandler)

	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles/", http.FileServer(http.Dir("./styles/"))))
	r.PathPrefix("/script.js").Handler(http.FileServer(http.Dir("./")))

	port := "8000"
	log.Printf("Server berjalan di port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func renderHTML(w http.ResponseWriter, tmpl string, data interface{}) {

	tmpl = fmt.Sprintf("templates/%s.html", tmpl)
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Periksa apakah user_id ada dalam session
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		// Jika belum login, redirect ke halaman login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Periksa apakah ada permintaan logout
		if r.FormValue("logout") == "true" {
			// Hapus session untuk melaksanakan logout
			session.Options.MaxAge = -1
			session.Save(r, w)

			// Redirect ke halaman login setelah logout berhasil
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	// Jika belum ada permintaan logout, tampilkan halaman home dengan data dari session
	data := map[string]interface{}{
		"FirstName": session.Values["first_name"],
		"LastName":  session.Values["last_name"],
	}
	renderHTML(w, "home", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "register-session")
	if err != nil {
		http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodPost {
		// Mendapatkan data dari formulir
		username := r.FormValue("username")
		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		password := r.FormValue("password")

		// Buat tabel users jika belum ada
		err := database.CreateUsersTable()
		if err != nil {
			http.Error(w, "Gagal membuat tabel users", http.StatusInternalServerError)
			return
		}

		// Simpan data ke database
		err = database.CreateUser(username, firstName, lastName, password)
		if err != nil {
			http.Error(w, "Gagal menyimpan data registrasi", http.StatusInternalServerError)
			return
		}
		session.Values["Message"] = "Registrasi berhasil! Silakan login dengan akun Anda."
		session.Save(r, w)

		// Redirect ke halaman login setelah registrasi berhasil
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Tampilkan halaman registrasi jika bukan metode POST
	renderHTML(w, "register", nil)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		// Mendapatkan data dari formulir
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Mengambil data pengguna dari database berdasarkan username dan password
		user, err := database.GetUserByUsernameAndPassword(username, password)
		if err != nil {
			http.Error(w, "Terjadi kesalahan saat login", http.StatusInternalServerError)
			return
		}

		if user == nil {
			// Jika login gagal, tampilkan pesan error
			http.Error(w, "Login gagal. Coba lagi.", http.StatusUnauthorized)
			return
		}

		// Jika login berhasil, atur session sebagai status login
		session, _ := store.Get(r, "session-login")
		session.Values["user_id"] = user.ID // Menggunakan ID pengguna yang diambil dari database
		session.Values["Message"] = "Login berhasil!"
		session.Save(r, w)

		// Redirect ke halaman home setelah login berhasil
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Tampilkan halaman login jika bukan metode POST
	renderHTML(w, "login", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Mendapatkan session dan user_id dari session
		session, _ := store.Get(r, "session-name")
		userID, ok := session.Values["user_id"].(int)
		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Hapus session untuk melaksanakan logout
		session.Options.MaxAge = -1
		session.Save(r, w)

		// Buat tabel logout_history jika belum ada
		err := database.CreateLogoutHistoryTable()
		if err != nil {
			log.Printf("Gagal membuat tabel logout_history: %v", err)
		} else {
			// Simpan riwayat logout ke dalam database (opsional, Anda bisa mengabaikannya jika tidak perlu)
			err = database.SaveLogoutHistory(userID)
			if err != nil {
				log.Printf("Gagal menyimpan riwayat logout ke database: %v", err)
			}
		}

		// Redirect ke halaman login setelah logout berhasil
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Tampilkan halaman konfirmasi logout jika bukan metode POST
	renderHTML(w, "logout", nil)
}

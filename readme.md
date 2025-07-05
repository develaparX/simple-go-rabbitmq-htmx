# Chat App - Golang Gin + RabbitMQ + HTMX

Aplikasi chat real-time untuk 2 pengguna dengan sistem autentikasi menggunakan Golang (Gin), RabbitMQ message broker, dan HTMX untuk interaksi frontend yang dinamis.

## Fitur

- ✅ **Autentikasi**: Login/logout dengan session management
- ✅ **Chat Real-time**: Pesan antar 2 pengguna
- ✅ **RabbitMQ Integration**: Message queue untuk reliabilitas
- ✅ **HTMX Interface**: UI dinamis tanpa JavaScript kompleks
- ✅ **Auto-refresh**: Update pesan setiap 3 detik
- ✅ **Responsive Design**: Interface modern dan mobile-friendly
- ✅ **Status Pesan**: Indikator pesan terbaca/belum
- ✅ **Enter to Send**: Kirim pesan dengan Enter

## Akun Demo

- **User 1**: `alice` / `password123`
- **User 2**: `bob` / `password123`

## Prasyarat

1. **Go** (versi 1.21 atau lebih baru)
2. **RabbitMQ** server

### Instalasi RabbitMQ

**Ubuntu/Debian:**

```bash
sudo apt update
sudo apt install rabbitmq-server
sudo systemctl start rabbitmq-server
sudo systemctl enable rabbitmq-server
```

**macOS (dengan Homebrew):**

```bash
brew install rabbitmq
brew services start rabbitmq
```

**Windows:**

- Download dari [rabbitmq.com](https://www.rabbitmq.com/download.html)
- Ikuti petunjuk instalasi

**Docker:**

```bash
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

## Instalasi dan Menjalankan

1. **Clone atau buat project:**

```bash
mkdir chat-app
cd chat-app
```

2. **Inisialisasi Go module:**

```bash
go mod init rabbitmq-gin-htmx
```

3. **Salin file-file:**

- `main.go` - File utama aplikasi
- `go.mod` - File dependensi
- Buat folder `templates/` dan salin file-file HTML:
  - `templates/login.html` - Halaman login
  - `templates/chat.html` - Halaman chat
  - `templates/messages.html` - Template pesan

4. **Install dependensi:**

```bash
go mod tidy
```

5. **Pastikan RabbitMQ berjalan:**

```bash
# Cek status RabbitMQ
sudo systemctl status rabbitmq-server

# Atau untuk Docker
docker ps | grep rabbitmq
```

6. **Jalankan aplikasi:**

```bash
go run main.go
```

7. **Akses aplikasi:**

- Buka browser ke `http://localhost:8080`
- Login dengan akun demo
- Untuk simulasi chat, buka browser kedua (atau incognito) dan login dengan akun yang berbeda

## Struktur Project

```
chat-app/
├── main.go
├── go.mod
├── go.sum
├── templates/
│   ├── login.html
│   ├── chat.html
│   └── messages.html
└── README.md
```

## Cara Kerja

### 1. Autentikasi

- User login dengan username/password
- Session disimpan menggunakan cookie-based sessions
- Middleware mengecek autentikasi untuk halaman terproteksi

### 2. Chat System

- Setiap pesan dikirim ke RabbitMQ queue
- Consumer goroutine mendengarkan pesan dari queue
- Interface HTMX melakukan auto-refresh setiap 3 detik
- Pesan ditampilkan dengan bubble chat yang berbeda untuk pengirim/penerima

### 3. RabbitMQ Flow

```
User A -> POST /send -> RabbitMQ Queue -> Consumer -> Log
User B -> GET /messages -> Tampilkan pesan User A
```

## Endpoints API

### Public Routes

- `GET /login` - Halaman login
- `POST /login` - Proses login
- `GET /logout` - Logout

### Protected Routes (require auth)

- `GET /` - Halaman chat utama
- `POST /send` - Kirim pesan
- `GET /messages` - Ambil daftar pesan
- `POST /mark-read/:id` - Tandai pesan sebagai dibaca

## Teknologi yang Digunakan

### Backend (Golang):

- **Gin Framework**: Web framework untuk routing dan middleware
- **Sessions**: Cookie-based session management
- **RabbitMQ**: Message broker untuk reliabilitas pesan
- **Goroutines**: Concurrent processing untuk consumer

### Frontend (HTMX):

- **hx-get/hx-post**: Ajax requests tanpa JavaScript
- **hx-trigger**: Auto-refresh dan event handling
- **hx-target/hx-swap**: Dynamic content updates
- **Modern CSS**: Responsive design dengan gradients dan animations

### Message Queue:

- **Publisher**: Mengirim pesan ke queue
- **Consumer**: Mendengarkan dan memproses pesan
- **JSON Serialization**: Struktur data pesan

## Fitur HTMX yang Digunakan

```html
<!-- Auto-refresh setiap 3 detik -->
<div hx-get="/messages" hx-trigger="every 3s">
  <!-- Submit form tanpa reload -->
  <form hx-post="/send" hx-target="#messages-list">
    <!-- Update konten dinamis -->
    <div hx-swap="innerHTML"></div>
  </form>
</div>
```

## Pengembangan Lebih Lanjut

### Level 1 (Beginner):

1. **Emoji Support**: Tambahkan emoji picker
2. **Typing Indicator**: Indikator sedang mengetik
3. **Message Timestamps**: Format waktu yang lebih baik
4. **Dark Mode**: Toggle tema gelap/terang

### Level 2 (Intermediate):

1. **Database**: PostgreSQL untuk persistent storage
2. **File Upload**: Kirim gambar/file
3. **Group Chat**: Chat untuk lebih dari 2 orang
4. **Push Notifications**: Browser notifications
5. **Message Search**: Pencarian dalam chat

### Level 3 (Advanced):

1. **WebSocket**: Real-time tanpa polling
2. **Microservices**: Pisahkan auth, chat, dan notification
3. **Redis**: Caching dan session store
4. **Docker Compose**: Multi-container deployment
5. **E2E Encryption**: Enkripsi pesan end-to-end

## Testing

### Manual Testing:

1. Buka 2 browser/tab berbeda
2. Login dengan alice di tab 1
3. Login dengan bob di tab 2
4. Kirim pesan bolak-balik
5. Verify auto-refresh bekerja

### Unit Testing (TODO):

```bash
go test ./...
```

## Troubleshooting

**❌ RabbitMQ connection refused:**

```bash
# Cek apakah RabbitMQ berjalan
sudo systemctl status rabbitmq-server

# Restart jika perlu
sudo systemctl restart rabbitmq-server
```

**❌ Session tidak bekerja:**

- Periksa cookie settings di browser
- Pastikan secret key di aplikasi aman
- Cek network tab di browser developer tools

**❌ HTMX tidak update:**

- Buka Console browser untuk cek error
- Pastikan CDN HTMX dapat diakses
- Verifikasi response dari server (Network tab)

**❌ Pesan tidak muncul:**

- Cek log aplikasi untuk error
- Pastikan RabbitMQ consumer berjalan
- Verifikasi JSON serialization

## Keamanan

⚠️ **Catatan**: Ini adalah aplikasi demo untuk pembelajaran. Untuk production:

1. Gunakan password hashing (bcrypt)
2. Implementasi CSRF protection
3. Validasi input yang ketat
4. Rate limiting untuk API
5. HTTPS untuk semua komunikasi
6. Environment variables untuk konfigurasi

## Lisensi

Untuk tujuan pembelajaran dan eksplorasi teknologi.

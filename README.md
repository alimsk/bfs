# Disclaimer
Pembuat bot ini tidak bertanggung jawab jika anda kena banned/blokir shopee

# Fitur
- Tidak mengandalkan webdriver atau selenium dapat membuat bot lebih cepat & ringan, dan support running di android via Termux
- Gratis ([donasi](#support))
- Multi akun
- Interactive CLI / [Simple](#simple-cli)

# Install
### Termux
Copas command dibawah
```
curl -L https://github.com/alimsk/bfs/releases/latest/download/bfs-android-arm64 > bfs && chmod +x bfs
```
gunakan command yang sama ketika mau update

### Linux
```
curl -L https://github.com/alimsk/bfs/releases/latest/download/bfs-linux-amd64 > bfs && chmod +x bfs
```

### Windows
Download file exe nya manual di [release](https://github.com/alimsk/bfs/releases/latest)

## Simple CLI
Kalo mau running botnya di vps/rdp, mungkin lebih enak menggunakan CLI yg simpel.  
untuk itu udah saya sediakan, bisa compile sendiri di folder bfs-simple (butuh go1.18).

Versi ini nggak ada bedanya dengan versi yg biasa, kecuali lebih simpel saja.

# Penggunaan
Untuk login bisa ambil cookie shopee dari chrome menggunakan ekstensi [Copy Cookies](https://chrome.google.com/webstore/detail/copy-cookies/jcbpglbplpblnagieibnemmkiamekcdg?hl=en),
lalu pastekan ke textinputnya.

Kalo kurang jelas bisa cek [video tutorial](https://youtu.be/1fIKouowm_M).

## CLI Arguments
### -state
nama state file.

state file adalah tempat bfs menyimpan akun dan data. (default "bfs_state.json")

### -d
delay antar request (pada bagian checkout).  
bot akan mengirimkan request secara bersamaan, opsi ini mengatur berapa lama harus menunggu sebelum mengirimkan request selanjutnya.

jika disetting ke 0 maka request akan dikirimkan satu-persatu.

nilai dapat berisi durasi seperti 1s, 500ms, atau 2m

### -sub
mengurangi waktu flash sale dengan nilai yg diberikan.  
misal waktu fs adalah 12:00:00, jika argumen ini 1s maka bot akan mulai checkout pada 11:59:59.  
karena 12:00:00 - 1s adalah 11:59:59.

Note: jika flash sale sudah dimulai maka opsi ini di abaikan.

nilai dapat berisi durasi seperti 1s, 500ms, atau 2m

## Subcommand
### info
mengambil informasi produk.

penggunaan:  
`bfs info <url produk>`

### version
tampilkan versi bfs.

penggunaan:  
`bfs version`

## Grup Telegram
https://t.me/+n9NofX8hLwo5OGVl

# Support
Star repo ini dengan menekan tombol <img align="center" src="https://user-images.githubusercontent.com/51353996/163677753-c95363d2-54aa-412e-8709-6daaf341223f.png">
 di pojok kanan atas.  
jangan cuma make doang dong, biar sama-sama untung :D

bisa juga donate ke saya via dana

<img src="https://user-images.githubusercontent.com/51353996/158705498-add7da42-1907-43ff-ab80-b2d673f66b3b.png" width="600">

kalo nemu bug atau masalah lain bisa buat issue di halaman [issue](https://github.com/alimsk/bfs/issues).

# Contribution
Contributions are welcome

# Author
[alimsk](https://github.com/alimsk)

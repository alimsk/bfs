# Disclaimer
Pembuat bot ini tidak bertanggung jawab jika anda kena banned/blokir shopee

# Fitur
- Tidak mengandalkan webdriver atau selenium dapat membuat bot lebih cepat & ringan, dan support running di android via Termux
- Gratis ([donasi](#support))
- Multi akun
- Interactive CLI

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

### Install dari source
Dibutuhkan go 1.18

install dari main branch untuk fitur terbaru (sebelum dirilis)
```
go install github.com/alimsk/bfs@main
```

# Penggunaan
Untuk login bisa ambil cookie shopee dari chrome menggunakan ekstensi [Copy Cookies](https://chrome.google.com/webstore/detail/copy-cookies/jcbpglbplpblnagieibnemmkiamekcdg?hl=en),
lalu pastekan ke textinputnya.

Kalo kurang jelas bisa cek [video tutorial](https://youtu.be/1fIKouowm_M).

## Arguments
### -state
nama state file.

state file adalah tempat bfs menyimpan akun dan data.

### -d
delay antar request (pada bagian checkout).
bot akan mengirimkan request secara bersamaan, opsi ini mengatur berapa lama harus menunggu sebelum mengirimkan request selanjutnya.

jika disetting ke 0 (default) maka request akan dikirimkan satu-persatu.

contoh:  
- 10ms
- 5s
- 2h

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
Star repo ini jika berguna, bisa juga donate ke saya via dana :)

<img src="https://user-images.githubusercontent.com/51353996/158705498-add7da42-1907-43ff-ab80-b2d673f66b3b.png" width="600">

kalo nemu bug atau masalah lain bisa buat issue di halaman [issue](https://github.com/alimsk/bfs/issues).

# Contribution
Contributions are welcome

# Author
[alimsk](https://github.com/alimsk)

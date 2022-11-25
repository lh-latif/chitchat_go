# ChitChat Go
Project ini saya buat 'from scratch' terinspirasi dari Whatsapp dan Telegram. Webserver chat ini menggunakan symmetric encryption yaitu Diffie-Hellman Key Exchange. Sehingga pesan yang dikirim secara teknis tidak dapat dibaca oleh pihak ketiga.

## Status
Project ini sudah bisa mengirimkan pesan antara 2 party. Namun karena ini masih early stage masih banyak yang harus diperbaiki dan refactor. Seperti _spagethi code_ di __main.go__, Dokumentasi Design architecture, Protocol Websocket, Interaksi antar _goroutine_, Unit Testing dan Integration Test.

## Vision
Saya sebenarnya berencana membuat chat server yang dapat scaling secara vertical dan modular, sehingga sebuah chat server dapat menggunakan multiple stack & tech. Saat ini saya sedang menulis menggunakan Go dan Erlang.
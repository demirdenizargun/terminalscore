# terminalscore

Terminalde Trendyol Süper Lig takibi: günün maçları, canlı skor, maç olayları ve puan durumu.
Veri ESPN'in herkese açık API'sinden gelir, API anahtarı gerekmez.

## Derleme

```sh
go build -o terminalscore .
./terminalscore
```

Windows'ta Linux için derlemek (PowerShell):

```powershell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o terminalscore-linux-amd64 .
```

## Tuşlar

| Tuş | İşlev |
|---|---|
| ← / → | Önceki / sonraki gün |
| t | Bugüne dön |
| ↑ / ↓ | Maç seç |
| enter | Maç detayı |
| esc | Geri |
| tab | Maçlar / Puan Durumu |
| r | Yenile |
| q | Çıkış |

Canlı maç varken skorlar 30 saniyede bir otomatik yenilenir.

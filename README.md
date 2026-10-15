# Happy Birthday Danh 🎂 — Roblox Birthday Server

🎮 **Live:** https://lolizabrett-byte.github.io/happy-birthday-danh/

Web chúc mừng sinh nhật tuổi 20 cho **Lưu Thành Danh** (Roblox: SaveAsDanh / @danhohmygod2), theme Roblox: Steal An Egg, Blox Fruits. From **kho**.

- **Frontend** (`site/`): HTML/CSS/ES modules thuần, không build step. Host trên GitHub Pages.
- **API** (`api/`): Go + Echo. Có sổ lưu bút và leaderboard trứng. Host trên Render free (Docker), dữ liệu lưu ở Turso (libSQL).
- **Asset tool** (`tools/fetch-assets/`): Go CLI tải avatar thật của Danh, icon game và ảnh trái ác quỷ về `site/assets/`.

Web vẫn chạy đầy đủ khi API tắt: sổ lưu bút chỉ hiện lời chúc mẫu, các game vẫn chơi bình thường.

## Tính năng
- Màn hình "Joining server…" kiểu Roblox. Bấm **Play** để mở nhạc Happy Birthday 8-bit (tự tổng hợp bằng WebAudio).
- Avatar thật của Danh lên **Lv 20**, có toast "Badge Awarded" và confetti.
- Profile card: bio thật và đồ đang mặc thật, gồm cả áo HCMUSSH.
- **Steal An Egg**: cướp trứng trên băng chuyền khi chủ base quay đi, ấp trứng để nhận lời chúc theo độ hiếm.
- **Blox Fruit Gacha**: roll trái ác quỷ, mỗi trái là một lời chúc, có túi sưu tầm.
- **Bánh kem**: thổi nến bằng cách bấm hoặc bằng micro.
- Lá thư từ kho hiện dạng chat Roblox.
- Sổ lưu bút và bảng xếp hạng (khi API bật).

## Chạy local
```bash
# Frontend
python -m http.server 8000 -d site        # mở http://localhost:8000

# API (memory store, không cần Turso)
cd api && go run .                         # :8080
go test -race ./...

# Tải lại asset (avatar, icon game, trái ác quỷ)
cd tools/fetch-assets && go run . -out ../../site/assets          # -force để tải đè
```
Muốn frontend gọi API local thì đặt `API_BASE = 'http://localhost:8080'` trong `site/js/config.js`. Mặc định CORS đã cho phép `http://localhost:8000`.

## Deploy
1. **GitHub Pages:** Pages publish nhánh `gh-pages`, nhánh này chứa nội dung thư mục `site/`. Sau khi commit lên `main`, chạy:
   ```bash
   git push origin main
   git subtree push --prefix site origin gh-pages
   ```
2. **Turso:** tạo database và token.
   ```bash
   turso db create danh-bd
   turso db show danh-bd --url
   turso db tokens create danh-bd
   ```
3. **Render:** vào *New → Blueprint*, chọn repo này (Render đọc `render.yaml`). Điền các env:
   | Env | Giá trị |
   |---|---|
   | `TURSO_DATABASE_URL` | `libsql://danh-bd-<org>.turso.io` |
   | `TURSO_AUTH_TOKEN` | token tạo ở bước 2 |
   | `ALLOWED_ORIGINS` | `https://lolizabrett-byte.github.io` (không kèm path repo, không có `/` ở cuối) |
4. Đặt URL của Render (`https://danh-birthday-api.onrender.com`) vào `API_BASE` trong `site/js/config.js`, rồi push.

Lưu ý: Render free sẽ ngủ sau 15 phút không có request. Trang web gọi `/healthz` ngay khi mở để đánh thức server, lần đầu mất khoảng 1 phút.

## Cấu trúc
```
site/                 frontend tĩnh (GitHub Pages)
  js/games/           steal-an-egg, fruit-gacha, cake, birthday-wishes
  assets/             ảnh và data do tools/fetch-assets sinh ra
api/                  Go Echo API (Dockerfile, tests)
tools/fetch-assets/   Go CLI tải asset
```

Fan-made, không liên kết với Roblox Corporation. Hình ảnh game thuộc về chủ sở hữu tương ứng.

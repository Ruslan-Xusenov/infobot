# Loyihani Git va Serverga Joylash Bo'yicha Qo'llanma

Ushbu qo'llanma loyihani GitHub-ga yuklash, uni **46.224.133.140** serveriga joylashtirish va avtomatik yangilanish skriptini (`deploy.sh`) sozlashni tushuntiradi.

---

## 1. Loyihani GitHub-ga Yuklash (Local kompyuterda)

Loyiha papkasida Git allaqachon sozlandi va remote URL `https://github.com/Ruslan-Xusenov/infobot.git` ga ulandi.

Fayllarni GitHub-ga yuklash uchun quyidagi buyruqlarni ketma-ket bajaring:

```bash
# 1. Barcha o'zgarishlarni git-ga qo'shish (.gitignore dagi fayllar kirmaydi)
git add .

# 2. O'zgarishlarni commit qilish
git commit -m "initial commit: deployment scripts added"

# 3. GitHub-ga yuklash (Push)
git push -u origin main
```

*Eslatma: Birinchi marta push qilganda GitHub username va token (yoki SSH kalit) so'ralishi mumkin.*

---

## 2. Serverda Sozlash (46.224.133.140 serverida)

Serverga SSH orqali ulaning va quyidagi bosqichlarni bajaring:

### A. Loyihani yuklab olish va Go o'rnatish
```bash
# Serverda Go va Git o'rnatilganligini tekshiring
sudo apt update
sudo apt install -y git golang-go

# Loyihani serverning /var/www papkasiga yuklab oling (yoki boshqa xohlagan papkaga)
sudo mkdir -p /var/www
sudo chown -R $USER:$USER /var/www
cd /var/www
git clone https://github.com/Ruslan-Xusenov/infobot.git
cd infobot
```

### B. Atrof-muhit (.env) faylini yaratish
`.env` fayli GitHub-ga yuklanmaydi (xavfsizlik uchun). Uni serverda qo'lda yaratish kerak:
```bash
cp .env.example .env
nano .env
```
Fayl ichidagi sozlamalarni serveringizga moslab tahrirlang (Bot token, Postgres ma'lumotlari).

### C. Systemd Servisini Sozlash
Bot orqa fonda (background) uzluksiz ishlashi va server o'chib yonganda avtomatik ishga tushishi uchun systemd servisidan foydalanamiz:

```bash
# 1. Servis faylini tizim papkasiga nusxalash
sudo cp infobot.service /etc/systemd/system/infobot.service

# 2. Tizim servislarini qayta yuklash
sudo systemctl daemon-reload

# 3. Servisni yoqish va ishga tushirish
sudo systemctl enable infobot --now

# 4. Servis holatini tekshirish
sudo systemctl status infobot
```

*Eslatma: Agar loyiha `/var/www/infobot` dan boshqa papkada bo'lsa yoki serverda `root` foydalanuvchisi ishlatilmasa, `/etc/systemd/system/infobot.service` ichidagi `WorkingDirectory`, `ExecStart` va `User` parametrlarini tahrirlang.*

---

## 3. Avtomatik Yangilash Skripti (`deploy.sh`)

Loyihaga `deploy.sh` skripti qo'shildi. Bu skript:
1. GitHub-dan yangi commitlar bor-yo'qligini tekshiradi (`git fetch`).
2. Agar yangilanish bo'lsa, kodni yangilaydi (`git reset --hard origin/main`), dasturni qaytadan build qiladi (`go build`) va botni qayta ishga tushiradi (`systemctl restart`).
3. Agar o'zgarish bo'lmasa, hech narsa qilmaydi va chiqib ketadi.

### Skriptni serverda ishga tushirish uchun:
1. Skriptga ishga tushirish ruxsatini bering (bir marta):
   ```bash
   chmod +x deploy.sh
   ```
2. Yangilamoqchi bo'lganingizda skriptni ishga tushiring:
   ```bash
   ./deploy.sh
   ```

### Avtomatlashtirish (Cron orqali):
Agar har 5 daqiqada skript o'zi avtomatik tekshirib yangilashini xohlasangiz, cron-ga qo'shishingiz mumkin:
```bash
crontab -e
```
Faylning oxiriga quyidagi qatorni qo'shing (yo'lni o'zingiznikiga moslang):
```text
*/5 * * * * /var/www/infobot/deploy.sh >> /var/www/infobot/deploy.log 2>&1
```
Bu har 5 daqiqada yangilanishlarni tekshiradi va loglarni `deploy.log` fayliga yozib boradi.

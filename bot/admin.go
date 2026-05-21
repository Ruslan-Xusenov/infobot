package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/company/infobot/database"
	"gopkg.in/telebot.v3"
)

var (
	adminMenu = &telebot.ReplyMarkup{}
	btnStat   = adminMenu.Data("📊 Analitika", "admin_stat")
	btnEdit   = adminMenu.Data("📝 Tugmalarni tahrirlash", "admin_edit")
	btnChan   = adminMenu.Data("📢 Majburiy obuna", "admin_chan")
	btnBcast  = adminMenu.Data("📨 Xabar tarqatish", "admin_bcast")
	btnAdmins = adminMenu.Data("👥 Adminlar", "admin_admins")
	btnExport = adminMenu.Data("📥 CSV Yuklab olish", "admin_export")
)

func RegisterAdminHandlers() {
	adminMenu.Inline(
		adminMenu.Row(btnStat, btnEdit),
		adminMenu.Row(btnChan, btnBcast),
		adminMenu.Row(btnAdmins, btnExport),
	)

	b.Handle("/admin", onAdmin)
	b.Handle(&btnStat, onAdminStat)
	b.Handle(&btnEdit, onAdminEditButtons)
	b.Handle(&btnChan, onAdminChannels)
	b.Handle(&btnBcast, onAdminBcast)
	b.Handle(&btnAdmins, onAdminAdmins)
	b.Handle(&btnExport, onAdminExport)
}

func onAdmin(c telebot.Context) error {
	if !database.IsAdmin(c.Sender().ID) {
		return c.Send("Sizda admin huquqi yo'q.")
	}
	ClearAdminState(c.Sender().ID)
	return c.Send("Admin paneliga xush kelibsiz. Nima qilamiz?", adminMenu)
}

func onAdminStat(c telebot.Context) error {
	stats, err := database.GetStats()
	if err != nil {
		return c.Send("Analitika olishda xatolik yuz berdi.")
	}

	text := fmt.Sprintf("📊 *Bot Analitikasi*\n\n"+
		"👥 Umumiy foydalanuvchilar: %d\n"+
		"✅ Faol foydalanuvchilar: %d\n"+
		"🚫 Bloklagan foydalanuvchilar: %d\n\n"+
		"📈 *O'sish dinamikasi:*\n"+
		"- Oxirgi 1 kunda: %d\n"+
		"- Oxirgi 1 haftada: %d\n"+
		"- Oxirgi 1 oyda: %d\n"+
		"- Oxirgi 3 oyda: %d",
		stats.TotalUsers, stats.ActiveUsers, stats.BlockedUsers,
		stats.JoinedToday, stats.JoinedWeek, stats.JoinedMonth, stats.Joined3Months)

	return c.Send(text, telebot.ModeMarkdown)
}

func onAdminEditButtons(c telebot.Context) error {
	menu := &telebot.ReplyMarkup{}
	var rows []telebot.Row
	btns := []string{"biz_kimmiz", "sotuv_bolimi", "shaxsiy_brend", "zapusk", "boglanish"}
	for _, bName := range btns {
		btn := menu.Data(bName, "edit_"+bName)
		b.Handle(&btn, func(cc telebot.Context) error {
			SetAdminState(cc.Sender().ID, "wait_content", map[string]interface{}{"button": cc.Callback().Unique[5:]})
			return cc.Send("Ushbu tugma uchun yangi matn, rasm, video yoki fayl yuboring. (Media fayllarga caption yozishingiz ham mumkin). Bekor qilish uchun /admin.")
		})
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return c.Send("Qaysi tugmani tahrirlaymiz?", menu)
}

func onAdminChannels(c telebot.Context) error {
	menu := &telebot.ReplyMarkup{}
	btnAdd := menu.Data("➕ Kanal qo'shish", "add_channel")
	b.Handle(&btnAdd, func(cc telebot.Context) error {
		SetAdminState(cc.Sender().ID, "wait_channel_info", nil)
		return cc.Send("Kanal ma'lumotlarini quyidagi formatda yuboring:\n`ChannelID URL Nomi`\nMasalan:\n`-100123456789 https://t.me/kanalim Mening Kanalim`\n\nBekor qilish uchun /admin", telebot.ModeMarkdown)
	})

	channels, _ := database.GetAllChannels()
	var rows []telebot.Row
	for _, ch := range channels {
		btnDel := menu.Data("❌ "+ch.Name, fmt.Sprintf("del_chan_%d", ch.ChannelID))
		chID := ch.ChannelID // Capture
		b.Handle(&btnDel, func(cc telebot.Context) error {
			database.DeleteChannel(chID)
			return cc.Send("Kanal o'chirildi. /admin")
		})
		rows = append(rows, menu.Row(btnDel))
	}
	rows = append(rows, menu.Row(btnAdd))
	menu.Inline(rows...)

	return c.Send("Majburiy obuna kanallarini boshqarish:", menu)
}

func onAdminBcast(c telebot.Context) error {
	SetAdminState(c.Sender().ID, "wait_broadcast", nil)
	return c.Send("Barcha foydalanuvchilarga tarqatish uchun xabar, rasm yoki video yuboring. Bekor qilish uchun /admin ni bosing.")
}

func onAdminAdmins(c telebot.Context) error {
	menu := &telebot.ReplyMarkup{}
	btnAdd := menu.Data("➕ Admin qo'shish", "add_admin")
	b.Handle(&btnAdd, func(cc telebot.Context) error {
		SetAdminState(cc.Sender().ID, "wait_admin_id", nil)
		return cc.Send("Yangi adminning Telegram ID raqamini yuboring:\nBekor qilish uchun /admin")
	})

	admins, _ := database.GetAllAdmins()
	var rows []telebot.Row
	for _, adminID := range admins {
		btnDel := menu.Data(fmt.Sprintf("❌ %d", adminID), fmt.Sprintf("del_adm_%d", adminID))
		aID := adminID // Capture
		b.Handle(&btnDel, func(cc telebot.Context) error {
			if aID == appConfig.AdminID {
				return cc.Send("Asosiy adminni o'chira olmaysiz!")
			}
			if aID == cc.Sender().ID {
				return cc.Send("Siz o'zingizni o'chira olmaysiz!")
			}
			database.RemoveAdmin(aID)
			return cc.Send("Admin o'chirildi. /admin")
		})
		rows = append(rows, menu.Row(btnDel))
	}
	rows = append(rows, menu.Row(btnAdd))
	menu.Inline(rows...)

	return c.Send("Adminlarni boshqarish:", menu)
}

func onAdminExport(c telebot.Context) error {
	users, err := database.GetAllUsers()
	if err != nil {
		return c.Send("Bazadan ma'lumotlarni olishda xatolik yuz berdi.")
	}

	var sb strings.Builder
	sb.WriteString("ID,Telegram ID,Ism,Username,Asosiy Raqam,Qo'shimcha Raqam,Status,Qo'shilgan sana\n")

	for _, u := range users {
		secPhone := ""
		if u.SecondaryPhone.Valid {
			secPhone = u.SecondaryPhone.String
		}

		sb.WriteString(fmt.Sprintf("%d,%d,%s,%s,%s,%s,%s,%s\n",
			u.ID, u.TelegramID,
			strings.ReplaceAll(u.FirstName, ",", " "),
			u.Username,
			u.PhoneNumber,
			secPhone,
			u.Status,
			u.CreatedAt.Format("2006-01-02 15:04"),
		))
	}

	doc := &telebot.Document{
		File:     telebot.FromReader(strings.NewReader(sb.String())),
		FileName: fmt.Sprintf("foydalanuvchilar_%s.csv", time.Now().Format("20060102")),
		Caption:  "📥 Barcha foydalanuvchilar ro'yxati (CSV formatida). Bu faylni Excel orqali ochishingiz mumkin.",
	}

	return c.Send(doc)
}

func onTextAdminCheck(c telebot.Context) error {
	if !database.IsAdmin(c.Sender().ID) {
		return nil
	}

	state := GetAdminState(c.Sender().ID)
	if state == nil {
		return nil // Not in state
	}

	text := c.Message().Text

	switch state.State {
	case "wait_content":
		btnName := state.Data["button"].(string)
		err := database.UpdateContent(btnName, text, "", "text")
		if err == nil {
			c.Send("✅ Tugma matni muvaffaqiyatli yangilandi.")
		}
		ClearAdminState(c.Sender().ID)

	case "wait_channel_info":
		parts := strings.SplitN(text, " ", 3)
		if len(parts) < 3 {
			return c.Send("Noto'g'ri format. Qaytadan yuboring:")
		}
		cID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return c.Send("Kanal ID raqam bo'lishi kerak. Qaytadan yuboring:")
		}
		err = database.AddChannel(cID, parts[1], parts[2])
		if err == nil {
			c.Send("✅ Kanal qo'shildi.")
		}
		ClearAdminState(c.Sender().ID)

	case "wait_admin_id":
		aID, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("ID raqam bo'lishi kerak. Qaytadan yuboring:")
		}
		err = database.AddAdmin(aID)
		if err == nil {
			c.Send("✅ Yangi admin qo'shildi.")
		}
		ClearAdminState(c.Sender().ID)

	case "wait_broadcast":
		go broadcastMessage(c.Message(), c.Sender().ID)
		ClearAdminState(c.Sender().ID)
		return c.Send("📨 Xabar tarqatish boshlandi...")
	}
	return nil
}

func onMediaAdminCheck(c telebot.Context) error {
	if !database.IsAdmin(c.Sender().ID) {
		return nil
	}

	state := GetAdminState(c.Sender().ID)
	if state == nil {
		return nil
	}

	msg := c.Message()

	if state.State == "wait_content" {
		btnName := state.Data["button"].(string)
		
		var fileID, mediaType string
		caption := msg.Caption

		if msg.Photo != nil {
			fileID = msg.Photo.FileID
			mediaType = "photo"
		} else if msg.Video != nil {
			fileID = msg.Video.FileID
			mediaType = "video"
		} else if msg.Voice != nil {
			fileID = msg.Voice.FileID
			mediaType = "voice"
		} else if msg.Document != nil {
			fileID = msg.Document.FileID
			mediaType = "document"
		}

		if fileID != "" {
			err := database.UpdateContent(btnName, caption, fileID, mediaType)
			if err == nil {
				c.Send("✅ Tugma kontenti muvaffaqiyatli yangilandi.")
			} else {
				c.Send("Xatolik yuz berdi.")
			}
		}
		ClearAdminState(c.Sender().ID)
		return nil
	}

	if state.State == "wait_broadcast" {
		go broadcastMessage(msg, c.Sender().ID)
		ClearAdminState(c.Sender().ID)
		return c.Send("📨 Xabar tarqatish boshlandi...")
	}

	return nil
}

func broadcastMessage(msg *telebot.Message, adminID int64) {
	ids, err := database.GetAllUsersTelegramIDs()
	if err != nil {
		log.Println("Broadcast xatosi:", err)
		return
	}

	sent, blocked := 0, 0
	for i, id := range ids {
		user := &telebot.User{ID: id}
		_, err := b.Copy(user, msg)
		if err != nil {
			// Check if bot was blocked
			if strings.Contains(err.Error(), "blocked") || strings.Contains(err.Error(), "deactivated") || strings.Contains(err.Error(), "not found") {
				database.UpdateUserStatus(id, "blocked")
				blocked++
			} else if strings.Contains(err.Error(), "Too Many Requests") {
				// Retry limits
				time.Sleep(3 * time.Second)
				b.Copy(user, msg) // Retry once
				sent++
			}
		} else {
			sent++
		}

		// Safe broadcasting logic (max ~25 msgs/sec overall)
		time.Sleep(40 * time.Millisecond)

		// Every 100 users, pause a little longer to reset Telegram's burst limit window
		if (i+1)%100 == 0 {
			time.Sleep(1 * time.Second)
		}
	}

	admin := &telebot.User{ID: adminID}
	b.Send(admin, fmt.Sprintf("✅ *Tarqatish tugadi!*\n\nMuvaffaqiyatli: %d\nBloklagan: %d", sent, blocked), telebot.ModeMarkdown)
}

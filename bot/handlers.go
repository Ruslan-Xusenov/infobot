package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/company/infobot/database"
	"gopkg.in/telebot.v3"
)

var callbackDataRx = regexp.MustCompile(`^\f([\w-]+)(\|(.*))?$`)

func buildMainMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}
	buttons, err := database.GetAllButtons()
	if err != nil || len(buttons) == 0 {
		return menu
	}
	var rows []telebot.Row
	for _, dbBtn := range buttons {
		btn := menu.Data(dbBtn.Label, dbBtn.UniqueName)
		rows = append(rows, menu.Row(btn))
	}
	menu.Inline(rows...)
	return menu
}

func RegisterHandlers() {
	b.Handle("/start", onStart)
	b.Handle("\fcheck_sub", onCheckSub)

	RegisterAdminHandlers()

	b.Handle(telebot.OnText, onText)
	b.Handle(telebot.OnMedia, onMediaAdminCheck)

	// Catch-all for dynamic menu button callbacks
	b.Handle(telebot.OnCallback, onDynamicMenuCallback)
}

const welcomeText = `🏢 *UySotPro Agentligi*

Biz quruvchilar uchun marketing, sotuv, shaxsiy brend va zapusk xizmatlarini kompleks tarzda ko'rsatamiz — va faqat *natija* uchun ishlaymiz.

👤 *Baxtishod Rasulov* — 5 yillik tajribaga ega marketolog va sotuv eksperti.

📍 Shu kunga qadar Samarqand, Surxondaryo va boshqa viloyatlardagi:
• City Park
• Oqsaroy Uylari
• Oltinsoy City
• Turon Uylari

loyihalarida yuzlab xonadonlarni sotib kelmoqdamiz.

👇 Xizmatlarimiz haqida to'liq ma'lumot olish uchun quyidagi tugmalardan birini tanlang:`

func onStart(c telebot.Context) error {
	user, err := database.GetUserByTelegramID(c.Sender().ID)
	if err != nil {
		log.Println("Error getting user:", err)
		return c.Send("Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
	}

	// Send welcome video + text to every user who starts
	sendWelcomeMessage(c)

	if user == nil || user.PhoneNumber == "" {
		removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
		return c.Send("📱 Botdan foydalanish va to'liq ma'lumot olish uchun telefon raqamingizni kiriting:\n\nFormat: *+998XXXXXXXXX*\n\nMasalan: +998901234567", removeKb, telebot.ModeMarkdown)
	}

	return checkAndSendMenu(c)
}

// sendWelcomeMessage sends the intro video (if configured) + welcome text
func sendWelcomeMessage(c telebot.Context) {
	content, err := database.GetContent("start_video")
	if err == nil && content != nil && content.MediaFileID != "" && content.MediaType == "video" {
		msg := &telebot.Video{
			File:    telebot.File{FileID: content.MediaFileID},
			Caption: welcomeText,
		}
		_ = c.Send(msg, telebot.ModeMarkdown)
	} else {
		_ = c.Send(welcomeText, telebot.ModeMarkdown)
	}
}

func onText(c telebot.Context) error {
	if GetAdminState(c.Sender().ID) != nil {
		return onTextAdminCheck(c)
	}

	user, err := database.GetUserByTelegramID(c.Sender().ID)
	if err != nil {
		return nil
	}

	phone := strings.TrimSpace(c.Message().Text)
	matched, _ := regexp.MatchString(`^\+998\d{9}$`, phone)

	// New user — save phone number
	if user == nil || user.PhoneNumber == "" {
		if !matched {
			return c.Send("❌ Noto'g'ri format! Iltimos, quyidagi formatda kiriting:\n\n*+998XXXXXXXXX*\n\nMasalan: +998901234567", telebot.ModeMarkdown)
		}
		if err2 := database.CreateUser(c.Sender().ID, c.Sender().FirstName, c.Sender().Username, phone); err2 != nil {
			log.Println("Error creating user:", err2)
			return c.Send("Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
		}
		c.Send("✅ Telefon raqam qabul qilindi. Rahmat!")
		return checkAndSendMenu(c)
	}

	return nil
}

func checkAndSendMenu(c telebot.Context) error {
	channels, err := database.GetAllChannels()
	if err != nil {
		log.Println("Error getting channels:", err)
		return c.Send("Xatolik yuz berdi.")
	}

	if len(channels) > 0 {
		var notSubscribed []database.Channel
		for _, ch := range channels {
			member, err := b.ChatMemberOf(telebot.ChatID(ch.ChannelID), c.Sender())
			if err != nil || member.Role == telebot.Left || member.Role == telebot.Kicked {
				notSubscribed = append(notSubscribed, ch)
			}
		}

		if len(notSubscribed) > 0 {
			menu := &telebot.ReplyMarkup{}
			var rows []telebot.Row
			for _, ch := range notSubscribed {
				btn := menu.URL("📢 "+ch.Name, ch.URL)
				rows = append(rows, menu.Row(btn))
			}
			btnCheck := menu.Data("✅ Obuna bo'ldim", "check_sub")
			rows = append(rows, menu.Row(btnCheck))
			menu.Inline(rows...)

			removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
			c.Send("Iltimos, botdan foydalanish uchun quyidagi kanallarga obuna bo'ling:", removeKb)
			return c.Send("Kanallarga obuna bo'lgach 'Obuna bo'ldim' tugmasini bosing:", menu)
		}
	}

	removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
	c.Send("Asosiy menyu:", removeKb)
	return c.Send("Bizning xizmatlar", buildMainMenu())
}

func onCheckSub(c telebot.Context) error {
	channels, err := database.GetAllChannels()
	if err != nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Xatolik yuz berdi!"})
	}

	for _, ch := range channels {
		member, err := b.ChatMemberOf(telebot.ChatID(ch.ChannelID), c.Sender())
		if err != nil || member.Role == telebot.Left || member.Role == telebot.Kicked {
			return c.Respond(&telebot.CallbackResponse{Text: "Hali hamma kanallarga obuna bo'lmadingiz!", ShowAlert: true})
		}
	}

	c.Respond(&telebot.CallbackResponse{Text: "Obuna tasdiqlandi!"})
	c.Delete()
	return checkAndSendMenu(c)
}

func onDynamicMenuCallback(c telebot.Context) error {
	// In telebot.v3, when a callback falls through to OnCallback,
	// c.Callback().Unique is always empty.
	// The raw data is in c.Callback().Data with \f prefix.
	rawData := c.Callback().Data
	log.Printf("Callback received: rawData=%q, sender=%d\n", rawData, c.Sender().ID)

	// Parse the \f-prefixed callback data to extract unique identifier
	unique := ""
	if len(rawData) > 0 && rawData[0] == '\f' {
		match := callbackDataRx.FindStringSubmatch(rawData)
		if match != nil {
			unique = match[1]
		}
	} else {
		unique = rawData
	}

	log.Printf("Parsed unique=%q\n", unique)

	if unique == "" {
		return c.Respond()
	}

	// ── Admin actions ──────────────────────────────────────────────
	if strings.HasPrefix(unique, "el_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:]
		SetAdminState(c.Sender().ID, "wait_edit_label", map[string]interface{}{"button": uname})
		c.Respond()
		return c.Send(fmt.Sprintf("«%s» tugmasining yangi nomini yuboring:\nBekor qilish: /admin", uname))
	}

	if strings.HasPrefix(unique, "ec_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:]
		SetAdminState(c.Sender().ID, "wait_content", map[string]interface{}{"button": uname})
		c.Respond()
		return c.Send("Yangi matn, rasm, video yoki fayl yuboring. (Caption yozishingiz mumkin). Bekor qilish: /admin")
	}

	if strings.HasPrefix(unique, "db_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:]
		confirmMenu := &telebot.ReplyMarkup{}
		btnYes := confirmMenu.Data("✅ Ha, o'chirish", "dbc_"+uname)
		btnNo := confirmMenu.Data("🔙 Bekor qilish", "admin_btns")
		confirmMenu.Inline(confirmMenu.Row(btnYes, btnNo))
		c.Respond()
		return c.Send(fmt.Sprintf("«%s» tugmasini o'chirishni tasdiqlaysizmi?", uname), confirmMenu)
	}

	if strings.HasPrefix(unique, "dbc_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		u := unique[4:]
		database.DeleteButton(u)
		c.Respond(&telebot.CallbackResponse{Text: "✅ Tugma o'chirildi!"})
		return sendButtonList(c)
	}

	// ── User menu button click ────────────────────────────────────
	btn, err := database.GetButton(unique)
	if err != nil {
		log.Printf("GetButton error for unique %q: %v\n", unique, err)
		return c.Respond()
	}
	if btn == nil {
		log.Printf("Button not found in database for unique %q\n", unique)
		return c.Respond()
	}

	content, err := database.GetContent(unique)
	if err != nil {
		log.Printf("GetContent error for unique %q: %v\n", unique, err)
		c.Respond()
		return c.Send("Kontent olishda xatolik yuz berdi: " + err.Error())
	}
	if content == nil {
		c.Respond()
		return c.Send("Kontent hali o'rnatilmagan.")
	}

	c.Respond()
	menu := buildMainMenu()

	switch content.MediaType {
	case "photo":
		msg := &telebot.Photo{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
		_, err = b.Send(c.Sender(), msg, menu)
	case "video":
		msg := &telebot.Video{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
		_, err = b.Send(c.Sender(), msg, menu)
	case "voice":
		msg := &telebot.Voice{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
		_, err = b.Send(c.Sender(), msg, menu)
	case "document":
		msg := &telebot.Document{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
		_, err = b.Send(c.Sender(), msg, menu)
	default:
		return c.Send(content.TextContent, menu)
	}
	return err
}


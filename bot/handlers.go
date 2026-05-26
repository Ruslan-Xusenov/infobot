package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/company/infobot/database"
	"gopkg.in/telebot.v3"
)

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
	b.Handle(telebot.OnContact, onContact)
	b.Handle("\fcheck_sub", onCheckSub)

	RegisterAdminHandlers()

	b.Handle(telebot.OnText, onText)
	b.Handle(telebot.OnMedia, onMediaAdminCheck)

	// Catch-all for dynamic menu button callbacks
	b.Handle(telebot.OnCallback, onDynamicMenuCallback)
}

func onStart(c telebot.Context) error {
	user, err := database.GetUserByTelegramID(c.Sender().ID)
	if err != nil {
		log.Println("Error getting user:", err)
		return c.Send("Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
	}

	if user == nil || user.PhoneNumber == "" {
		menu := &telebot.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
		btnContact := menu.Contact("📱 Telefon raqamni yuborish")
		menu.Reply(menu.Row(btnContact))
		return c.Send("Assalomu alaykum! Botdan to'liq foydalanish uchun asosiy telefon raqamingizni yuboring:", menu)
	}

	if !user.SecondaryPhone.Valid || user.SecondaryPhone.String == "" {
		removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
		return c.Send("Iltimos, qo'shimcha yana bitta telefon raqamingizni yozib yuboring (masalan: +998901234567):", removeKb)
	}

	return checkAndSendMenu(c)
}

func onContact(c telebot.Context) error {
	contact := c.Message().Contact
	if contact == nil || contact.UserID != c.Sender().ID {
		return c.Send("Iltimos, o'zingizning telefon raqamingizni yuboring!")
	}

	err := database.CreateUser(c.Sender().ID, c.Sender().FirstName, c.Sender().Username, contact.PhoneNumber)
	if err != nil {
		log.Println("Error creating user:", err)
		return c.Send("Xatolik yuz berdi. Iltimos, qayta urinib ko'ring.")
	}

	removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
	return c.Send("✅ Asosiy raqam qabul qilindi.\n\nIltimos, qo'shimcha yana bitta telefon raqamingizni yozib yuboring (masalan: +998901234567):", removeKb)
}

func onText(c telebot.Context) error {
	if GetAdminState(c.Sender().ID) != nil {
		return onTextAdminCheck(c)
	}

	user, err := database.GetUserByTelegramID(c.Sender().ID)
	if err != nil || user == nil {
		return nil
	}

	if user.PhoneNumber != "" && (!user.SecondaryPhone.Valid || user.SecondaryPhone.String == "") {
		phone := strings.TrimSpace(c.Message().Text)
		matched, _ := regexp.MatchString(`^\+998\d{9}$`, phone)
		if !matched {
			return c.Send("Noto'g'ri format! Iltimos, faqat O'zbekiston raqamini quyidagi formatda yozing:\n+998901234567")
		}

		database.UpdateUserSecondaryPhone(c.Sender().ID, phone)
		c.Send("✅ Qo'shimcha raqam qabul qilindi.")
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
	unique := c.Callback().Unique

	// Check if it is an admin action
	if strings.HasPrefix(unique, "el_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:] // strip "el_"
		SetAdminState(c.Sender().ID, "wait_edit_label", map[string]interface{}{"button": uname})
		c.Respond()
		return c.Send(fmt.Sprintf("«%s» tugmasining yangi nomini yuboring:\nBekor qilish: /admin", uname))
	}

	if strings.HasPrefix(unique, "ec_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:] // strip "ec_"
		SetAdminState(c.Sender().ID, "wait_content", map[string]interface{}{"button": uname})
		c.Respond()
		return c.Send("Yangi matn, rasm, video yoki fayl yuboring. (Caption yozishingiz mumkin). Bekor qilish: /admin")
	}

	if strings.HasPrefix(unique, "db_") {
		if !database.IsAdmin(c.Sender().ID) {
			return c.Respond()
		}
		uname := unique[3:] // strip "db_"
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
		u := unique[4:] // strip "dbc_"
		database.DeleteButton(u)
		c.Respond(&telebot.CallbackResponse{Text: "✅ Tugma o'chirildi!"})
		return sendButtonList(c)
	}

	// Default: dynamic menu button click by user
	btn, err := database.GetButton(unique)
	if err != nil || btn == nil {
		return c.Respond()
	}

	content, err := database.GetContent(unique)
	if err != nil {
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

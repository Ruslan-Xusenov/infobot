package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/company/infobot/database"
	"gopkg.in/telebot.v3"
)

var (
	mainMenu = &telebot.ReplyMarkup{}
	btnBizKimmiz = mainMenu.Data("Biz kimmiz", "biz_kimmiz")
	btnSotuv     = mainMenu.Data("Sotuv boʻlimi", "sotuv_bolimi")
	btnShaxsiy   = mainMenu.Data("Shaxsiy brend", "shaxsiy_brend")
	btnZapusk    = mainMenu.Data("Zapusk", "zapusk")
	btnBoglanish = mainMenu.Data("Bogʻlanish", "boglanish")
)

func RegisterHandlers() {
	mainMenu.Inline(
		mainMenu.Row(btnBizKimmiz, btnSotuv),
		mainMenu.Row(btnShaxsiy, btnZapusk),
		mainMenu.Row(btnBoglanish),
	)

	b.Handle("/start", onStart)
	b.Handle(telebot.OnContact, onContact)
	b.Handle("\fcheck_sub", onCheckSub)
	
	// Menu callbacks
	b.Handle(&btnBizKimmiz, onMenuCallback)
	b.Handle(&btnSotuv, onMenuCallback)
	b.Handle(&btnShaxsiy, onMenuCallback)
	b.Handle(&btnZapusk, onMenuCallback)
	b.Handle(&btnBoglanish, onMenuCallback)

	RegisterAdminHandlers()
	
	b.Handle(telebot.OnText, onText)
	b.Handle(telebot.OnMedia, onMediaAdminCheck)
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
				btn := menu.URL(fmt.Sprintf("📢 %s", ch.Name), ch.URL)
				rows = append(rows, menu.Row(btn))
			}
			btnCheck := menu.Data("✅ Obuna bo'ldim", "check_sub")
			rows = append(rows, menu.Row(btnCheck))
			menu.Inline(rows...)
			
			// Remove contact keyboard
			removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
			c.Send("Iltimos, botdan foydalanish uchun quyidagi kanallarga obuna bo'ling:", removeKb)
			return c.Send("Kanallarga obuna bo'lgach 'Obuna bo'ldim' tugmasini bosing:", menu)
		}
	}

	removeKb := &telebot.ReplyMarkup{RemoveKeyboard: true}
	c.Send("Asosiy menyu:", removeKb)
	return c.Send("Iltimos, kerakli bo'limni tanlang:", mainMenu)
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
	c.Delete() // Delete the sub message
	return checkAndSendMenu(c)
}

func onMenuCallback(c telebot.Context) error {
	buttonName := c.Callback().Unique
	content, err := database.GetContent(buttonName)
	if err != nil {
		log.Println("Error getting content:", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ma'lumot topilmadi!"})
	}

	c.Respond()

	var msg interface{}
	switch content.MediaType {
	case "photo":
		msg = &telebot.Photo{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
	case "video":
		msg = &telebot.Video{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
	case "voice":
		msg = &telebot.Voice{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
	case "document":
		msg = &telebot.Document{File: telebot.File{FileID: content.MediaFileID}, Caption: content.TextContent}
	default: // text
		return c.Send(content.TextContent, mainMenu)
	}

	_, err = b.Send(c.Sender(), msg, mainMenu)
	return err
}

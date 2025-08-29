package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// KeyboardBuilder helps build Telegram inline keyboards
type KeyboardBuilder struct {
	rows [][]tgbotapi.InlineKeyboardButton
}

// NewKeyboardBuilder creates a new keyboard builder
func NewKeyboardBuilder() *KeyboardBuilder {
	return &KeyboardBuilder{
		rows: make([][]tgbotapi.InlineKeyboardButton, 0),
	}
}

// AddRow adds a new row to the keyboard
func (kb *KeyboardBuilder) AddRow(buttons ...tgbotapi.InlineKeyboardButton) *KeyboardBuilder {
	kb.rows = append(kb.rows, buttons)
	return kb
}

// AddButton adds a single button to the last row
func (kb *KeyboardBuilder) AddButton(button tgbotapi.InlineKeyboardButton) *KeyboardBuilder {
	if len(kb.rows) == 0 {
		kb.rows = append(kb.rows, []tgbotapi.InlineKeyboardButton{})
	}
	lastRowIndex := len(kb.rows) - 1
	kb.rows[lastRowIndex] = append(kb.rows[lastRowIndex], button)
	return kb
}

// Build creates the final keyboard markup
func (kb *KeyboardBuilder) Build() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(kb.rows...)
}

// CreateMainKeyboard creates the main menu keyboard
func CreateMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Get Free Trial", "trial"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ My Account", "account"),
		).
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "help"),
		).
		Build()
}

// CreateAccountKeyboard creates the account management keyboard
func CreateAccountKeyboard() tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Usage Stats", "usage"),
			tgbotapi.NewInlineKeyboardButtonData("🔧 Settings", "settings"),
		).
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Main", "main"),
		).
		Build()
}

// CreateHelpKeyboard creates the help keyboard
func CreateHelpKeyboard() tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("📖 FAQ", "faq"),
			tgbotapi.NewInlineKeyboardButtonData("📞 Support", "support"),
		).
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Main", "main"),
		).
		Build()
}

// CreateTrialKeyboard creates the trial activation keyboard
func CreateTrialKeyboard() tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Activate Trial", "activate_trial"),
		).
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Main", "main"),
		).
		Build()
}

// CreateConfirmationKeyboard creates a confirmation keyboard
func CreateConfirmationKeyboard(action string) tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Yes", "confirm_"+action),
			tgbotapi.NewInlineKeyboardButtonData("❌ No", "cancel_"+action),
		).
		Build()
}

// CreateBackKeyboard creates a simple back button keyboard
func CreateBackKeyboard(backAction string) tgbotapi.InlineKeyboardMarkup {
	return NewKeyboardBuilder().
		AddRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", backAction),
		).
		Build()
}

// CreateEmptyKeyboard creates an empty keyboard (useful for removing keyboards)
func CreateEmptyKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup()
}

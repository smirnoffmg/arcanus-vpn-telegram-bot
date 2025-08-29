package utils

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewKeyboardBuilder(t *testing.T) {
	kb := NewKeyboardBuilder()
	assert.NotNil(t, kb)
	assert.Empty(t, kb.rows)
}

func TestKeyboardBuilder_AddRow(t *testing.T) {
	kb := NewKeyboardBuilder()
	button1 := tgbotapi.NewInlineKeyboardButtonData("Test1", "test1")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Test2", "test2")

	kb.AddRow(button1, button2)

	assert.Len(t, kb.rows, 1)
	assert.Len(t, kb.rows[0], 2)
	assert.Equal(t, "Test1", kb.rows[0][0].Text)
	assert.Equal(t, "test1", *kb.rows[0][0].CallbackData)
	assert.Equal(t, "Test2", kb.rows[0][1].Text)
	assert.Equal(t, "test2", *kb.rows[0][1].CallbackData)
}

func TestKeyboardBuilder_AddButton(t *testing.T) {
	kb := NewKeyboardBuilder()
	button1 := tgbotapi.NewInlineKeyboardButtonData("Test1", "test1")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Test2", "test2")

	kb.AddButton(button1).AddButton(button2)

	assert.Len(t, kb.rows, 1)
	assert.Len(t, kb.rows[0], 2)
	assert.Equal(t, "Test1", kb.rows[0][0].Text)
	assert.Equal(t, "Test2", kb.rows[0][1].Text)
}

func TestKeyboardBuilder_Build(t *testing.T) {
	kb := NewKeyboardBuilder()
	button1 := tgbotapi.NewInlineKeyboardButtonData("Test1", "test1")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Test2", "test2")

	kb.AddRow(button1).AddRow(button2)

	keyboard := kb.Build()
	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
	assert.Len(t, keyboard.InlineKeyboard[0], 1)
	assert.Len(t, keyboard.InlineKeyboard[1], 1)
}

func TestCreateMainKeyboard(t *testing.T) {
	keyboard := CreateMainKeyboard()

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)

	// First row should have 2 buttons
	assert.Len(t, keyboard.InlineKeyboard[0], 2)
	assert.Equal(t, "🔑 Get Free Trial", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, "trial", *keyboard.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "⚙️ My Account", keyboard.InlineKeyboard[0][1].Text)
	assert.Equal(t, "account", *keyboard.InlineKeyboard[0][1].CallbackData)

	// Second row should have 1 button
	assert.Len(t, keyboard.InlineKeyboard[1], 1)
	assert.Equal(t, "❓ Help", keyboard.InlineKeyboard[1][0].Text)
	assert.Equal(t, "help", *keyboard.InlineKeyboard[1][0].CallbackData)
}

func TestCreateAccountKeyboard(t *testing.T) {
	keyboard := CreateAccountKeyboard()

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)

	// First row should have 2 buttons
	assert.Len(t, keyboard.InlineKeyboard[0], 2)
	assert.Equal(t, "📊 Usage Stats", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, "usage", *keyboard.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "🔧 Settings", keyboard.InlineKeyboard[0][1].Text)
	assert.Equal(t, "settings", *keyboard.InlineKeyboard[0][1].CallbackData)

	// Second row should have 1 button
	assert.Len(t, keyboard.InlineKeyboard[1], 1)
	assert.Equal(t, "⬅️ Back to Main", keyboard.InlineKeyboard[1][0].Text)
	assert.Equal(t, "main", *keyboard.InlineKeyboard[1][0].CallbackData)
}

func TestCreateHelpKeyboard(t *testing.T) {
	keyboard := CreateHelpKeyboard()

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)

	// First row should have 2 buttons
	assert.Len(t, keyboard.InlineKeyboard[0], 2)
	assert.Equal(t, "📖 FAQ", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, "faq", *keyboard.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "📞 Support", keyboard.InlineKeyboard[0][1].Text)
	assert.Equal(t, "support", *keyboard.InlineKeyboard[0][1].CallbackData)

	// Second row should have 1 button
	assert.Len(t, keyboard.InlineKeyboard[1], 1)
	assert.Equal(t, "⬅️ Back to Main", keyboard.InlineKeyboard[1][0].Text)
	assert.Equal(t, "main", *keyboard.InlineKeyboard[1][0].CallbackData)
}

func TestCreateTrialKeyboard(t *testing.T) {
	keyboard := CreateTrialKeyboard()

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)

	// First row should have 1 button
	assert.Len(t, keyboard.InlineKeyboard[0], 1)
	assert.Equal(t, "✅ Activate Trial", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, "activate_trial", *keyboard.InlineKeyboard[0][0].CallbackData)

	// Second row should have 1 button
	assert.Len(t, keyboard.InlineKeyboard[1], 1)
	assert.Equal(t, "⬅️ Back to Main", keyboard.InlineKeyboard[1][0].Text)
	assert.Equal(t, "main", *keyboard.InlineKeyboard[1][0].CallbackData)
}

func TestCreateConfirmationKeyboard(t *testing.T) {
	action := "delete_account"
	keyboard := CreateConfirmationKeyboard(action)

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
	assert.Len(t, keyboard.InlineKeyboard[0], 2)

	assert.Equal(t, "✅ Yes", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, "confirm_delete_account", *keyboard.InlineKeyboard[0][0].CallbackData)
	assert.Equal(t, "❌ No", keyboard.InlineKeyboard[0][1].Text)
	assert.Equal(t, "cancel_delete_account", *keyboard.InlineKeyboard[0][1].CallbackData)
}

func TestCreateBackKeyboard(t *testing.T) {
	backAction := "main_menu"
	keyboard := CreateBackKeyboard(backAction)

	assert.NotNil(t, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
	assert.Len(t, keyboard.InlineKeyboard[0], 1)

	assert.Equal(t, "⬅️ Back", keyboard.InlineKeyboard[0][0].Text)
	assert.Equal(t, backAction, *keyboard.InlineKeyboard[0][0].CallbackData)
}

func TestCreateEmptyKeyboard(t *testing.T) {
	keyboard := CreateEmptyKeyboard()

	assert.NotNil(t, keyboard)
	assert.Empty(t, keyboard.InlineKeyboard)
}

func TestKeyboardBuilder_Chaining(t *testing.T) {
	kb := NewKeyboardBuilder()
	button1 := tgbotapi.NewInlineKeyboardButtonData("Test1", "test1")
	button2 := tgbotapi.NewInlineKeyboardButtonData("Test2", "test2")
	button3 := tgbotapi.NewInlineKeyboardButtonData("Test3", "test3")

	// Test method chaining
	kb.AddRow(button1, button2).AddRow(button3)

	assert.Len(t, kb.rows, 2)
	assert.Len(t, kb.rows[0], 2)
	assert.Len(t, kb.rows[1], 1)

	assert.Equal(t, "Test1", kb.rows[0][0].Text)
	assert.Equal(t, "Test2", kb.rows[0][1].Text)
	assert.Equal(t, "Test3", kb.rows[1][0].Text)
}

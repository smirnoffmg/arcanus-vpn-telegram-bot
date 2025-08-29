package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/events"
)

// BotAPI interface for Telegram bot operations
type BotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	StopReceivingUpdates()
	GetMe() (tgbotapi.User, error)
}

// Handler handles Telegram bot interactions
type Handler struct {
	botAPI       BotAPI
	userService  domain.UserService
	logger       *logrus.Logger
	rateLimiter  *RateLimiter
	auditLogger  *AuditLogger
	processLock  *ProcessLock
	eventService *events.Service
}

// NewHandler creates a new bot handler
func NewHandler(botAPI BotAPI, userService domain.UserService, logger *logrus.Logger) *Handler {
	return &Handler{
		botAPI:      botAPI,
		userService: userService,
		logger:      logger,
		rateLimiter: NewRateLimiter(),
		auditLogger: NewAuditLogger(logger),
		processLock: NewProcessLock(""),
	}
}

// NewHandlerWithEvents creates a new bot handler with event publishing
func NewHandlerWithEvents(botAPI BotAPI, userService domain.UserService, logger *logrus.Logger, eventService *events.Service) *Handler {
	return &Handler{
		botAPI:       botAPI,
		userService:  userService,
		logger:       logger,
		rateLimiter:  NewRateLimiter(),
		auditLogger:  NewAuditLogger(logger),
		processLock:  NewProcessLock(""),
		eventService: eventService,
	}
}

// HandleUpdate processes incoming Telegram updates
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	message := update.Message
	h.logger.WithFields(logrus.Fields{
		"chat_id":    message.Chat.ID,
		"user_id":    message.From.ID,
		"username":   message.From.UserName,
		"text":       message.Text,
		"message_id": message.MessageID,
	}).Info("Received message")

	// Publish bot message received event
	if h.eventService != nil {
		command := message.Command()
		if err := h.eventService.PublishBotMessageReceived(ctx, message.From.ID, message.From.UserName, message.Chat.ID, message.MessageID, message.Text, command); err != nil {
			h.logger.WithError(err).Error("Failed to publish bot message received event")
		}
	}

	switch message.Text {
	case "/start":
		return h.handleStart(ctx, message)
	case "/account":
		return h.handleAccount(ctx, message)
	case "/help":
		return h.handleHelp(ctx, message)
	default:
		return h.handleUnknownCommand(ctx, message)
	}
}

// HandleCallback handles inline keyboard callbacks
func (h *Handler) HandleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	h.logger.WithFields(logrus.Fields{
		"chat_id":    callback.Message.Chat.ID,
		"user_id":    callback.From.ID,
		"username":   callback.From.UserName,
		"data":       callback.Data,
		"message_id": callback.Message.MessageID,
	}).Info("Received callback")

	// Publish bot callback received event
	if h.eventService != nil {
		if err := h.eventService.PublishBotCallbackReceived(ctx, callback.From.ID, callback.From.UserName, callback.Message.Chat.ID, callback.Message.MessageID, callback.Data); err != nil {
			h.logger.WithError(err).Error("Failed to publish bot callback received event")
		}
	}

	switch callback.Data {
	case "trial":
		return h.handleTrialActivation(ctx, callback)
	case "account":
		return h.handleAccountCallback(ctx, callback)
	case "help":
		return h.handleHelpCallback(ctx, callback)
	default:
		return h.handleUnknownCallback(ctx, callback)
	}
}

// handleStart handles the /start command
func (h *Handler) handleStart(ctx context.Context, message *tgbotapi.Message) error {
	user, err := h.userService.RegisterUser(ctx, message.From.ID, message.From.UserName, message.From.FirstName, message.From.LastName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to register user")
		return h.sendErrorMessage(message.Chat.ID, "Failed to register user. Please try again.")
	}

	text := fmt.Sprintf("🎉 Welcome to Arcanus VPN, %s!\n\n"+
		"🔐 Secure, private, and fast VPN service\n"+
		"📊 You have %s of free trial data\n\n"+
		"Choose an option below:",
		user.FirstName,
		formatBytes(user.QuotaLimit))

	keyboard := h.createMainKeyboard()
	return h.sendMessage(message.Chat.ID, text, keyboard)
}

// handleAccount handles the /account command
func (h *Handler) handleAccount(ctx context.Context, message *tgbotapi.Message) error {
	user, err := h.userService.GetUser(ctx, message.From.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		return h.sendErrorMessage(message.Chat.ID, "Failed to get account information. Please try again.")
	}

	text := h.formatAccountInfo(user)
	keyboard := h.createMainKeyboard()
	return h.sendMessage(message.Chat.ID, text, keyboard)
}

// handleHelp handles the /help command
func (h *Handler) handleHelp(ctx context.Context, message *tgbotapi.Message) error {
	text := `🤖 **Arcanus VPN Bot Help**

**Commands:**
• /start - Register and get started
• /account - View your account details
• /help - Show this help message

**Features:**
• 🔐 Secure VPN connection
• 📊 50MB free trial
• ⚡ Fast and reliable
• 🛡️ Privacy-focused

**Support:**
For technical support, contact @support`

	keyboard := h.createMainKeyboard()
	return h.sendMessage(message.Chat.ID, text, keyboard)
}

// handleUnknownCommand handles unknown commands
func (h *Handler) handleUnknownCommand(ctx context.Context, message *tgbotapi.Message) error {
	text := "❓ Unknown command. Use /help to see available commands."
	keyboard := h.createMainKeyboard()
	return h.sendMessage(message.Chat.ID, text, keyboard)
}

// handleTrialActivation handles trial activation callback
func (h *Handler) handleTrialActivation(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	err := h.userService.ActivateTrial(ctx, callback.From.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to activate trial")
		return h.answerCallback(callback.ID, "❌ Failed to activate trial. Please try again.")
	}

	user, err := h.userService.GetUser(ctx, callback.From.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user after trial activation")
		return h.answerCallback(callback.ID, "✅ Trial activated! But failed to get account details.")
	}

	text := fmt.Sprintf("🎉 **Trial Activated!**\n\n"+
		"Your account is now active with %s of data.\n"+
		"Enjoy secure browsing!",
		formatBytes(user.QuotaLimit))

	keyboard := h.createMainKeyboard()
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
}

// handleAccountCallback handles account callback
func (h *Handler) handleAccountCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	user, err := h.userService.GetUser(ctx, callback.From.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		return h.answerCallback(callback.ID, "❌ Failed to get account information.")
	}

	text := h.formatAccountInfo(user)
	keyboard := h.createMainKeyboard()
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
}

// handleHelpCallback handles help callback
func (h *Handler) handleHelpCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	text := `🤖 **Arcanus VPN Bot Help**

**Commands:**
• /start - Register and get started
• /account - View your account details
• /help - Show this help message

**Features:**
• 🔐 Secure VPN connection
• 📊 50MB free trial
• ⚡ Fast and reliable
• 🛡️ Privacy-focused

**Support:**
For technical support, contact @support`

	keyboard := h.createMainKeyboard()
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, text, keyboard)
}

// handleUnknownCallback handles unknown callbacks
func (h *Handler) handleUnknownCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	return h.answerCallback(callback.ID, "❓ Unknown action. Please try again.")
}

// createMainKeyboard creates the main inline keyboard
func (h *Handler) createMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Get Free Trial", "trial"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ My Account", "account"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "help"),
		),
	)
}

// formatAccountInfo formats user account information
func (h *Handler) formatAccountInfo(user *domain.User) string {
	status := "🔴 Inactive"
	if user.IsActive() {
		status = "🟢 Active"
	}

	// Escape underscores in username for Markdown
	escapedUsername := strings.ReplaceAll(user.Username, "_", "\\_")

	return fmt.Sprintf("📊 **Account Information**\n\n"+
		"👤 **Name:** %s %s\n"+
		"🆔 **Username:** @%s\n"+
		"📈 **Status:** %s\n"+
		"💾 **Data Limit:** %s\n"+
		"📊 **Data Used:** %s\n"+
		"📋 **Data Remaining:** %s\n"+
		"📅 **Member Since:** %s",
		user.FirstName, user.LastName,
		escapedUsername,
		status,
		formatBytes(user.QuotaLimit),
		formatBytes(user.QuotaUsed),
		formatBytes(user.GetQuotaRemaining()),
		user.CreatedAt.Format("Jan 2, 2006"))
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// sendMessage sends a message with optional keyboard
func (h *Handler) sendMessage(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := h.botAPI.Send(msg)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"chat_id": chatID,
			"text":    text,
		}).Error("Failed to send message")
		return fmt.Errorf("failed to send message: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"chat_id": chatID,
		"text":    text,
	}).Info("Message sent successfully")

	return nil
}

// sendErrorMessage sends an error message
func (h *Handler) sendErrorMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err := h.botAPI.Send(msg)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"chat_id": chatID,
			"text":    text,
		}).Error("Failed to send error message")
		return fmt.Errorf("failed to send error message: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"chat_id": chatID,
		"text":    text,
	}).Info("Error message sent successfully")

	return nil
}

// editMessage edits an existing message
func (h *Handler) editMessage(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard

	_, err := h.botAPI.Send(edit)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"chat_id":    chatID,
			"message_id": messageID,
			"text":       text,
		}).Error("Failed to edit message")
		return fmt.Errorf("failed to edit message: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	}).Info("Message edited successfully")

	return nil
}

// answerCallback answers a callback query
func (h *Handler) answerCallback(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := h.botAPI.Request(callback)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"callback_id": callbackID,
			"text":        text,
		}).Error("Failed to answer callback")
		return fmt.Errorf("failed to answer callback: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"callback_id": callbackID,
		"text":        text,
	}).Info("Callback answered successfully")

	return nil
}

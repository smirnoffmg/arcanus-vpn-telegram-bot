package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/domain"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/middleware"
	"github.com/smirnoffmg/arcanus-vpn-telegram-bot/internal/utils"
)

// HandlerWithMiddleware is a bot handler that uses middleware
type HandlerWithMiddleware struct {
	botAPI         BotAPI
	userService    domain.UserService
	logger         *logrus.Logger
	messageHandler middleware.HandlerFunc
	callbackHandler middleware.HandlerFunc
}

// NewHandlerWithMiddleware creates a new middleware-aware handler
func NewHandlerWithMiddleware(
	botAPI BotAPI,
	userService domain.UserService,
	logger *logrus.Logger,
	rateLimiter *RateLimiter,
	auditLogger *AuditLogger,
) *HandlerWithMiddleware {
	h := &HandlerWithMiddleware{
		botAPI:      botAPI,
		userService: userService,
		logger:      logger,
	}

	// Create middleware
	rateLimiterAdapter := NewRateLimiterAdapter(rateLimiter)
	auditLoggerAdapter := NewAuditLoggerAdapter(auditLogger)

	// Chain middleware for message handling
	h.messageHandler = middleware.Chain(
		h.handleMessageWithMiddleware,
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.Timeout(30*time.Second),
		middleware.RateLimit(rateLimiterAdapter),
		middleware.Audit(auditLoggerAdapter),
	)

	// Chain middleware for callback handling
	h.callbackHandler = middleware.Chain(
		h.handleCallbackWithMiddleware,
		middleware.Logger(logger),
		middleware.Recovery(logger),
		middleware.Timeout(30*time.Second),
		middleware.RateLimit(rateLimiterAdapter),
		middleware.Audit(auditLoggerAdapter),
	)

	return h
}

// HandleUpdate handles incoming Telegram updates using middleware
func (h *HandlerWithMiddleware) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message != nil {
		requestData := middleware.NewRequestDataFromUpdate(&update)
		return h.messageHandler(ctx, requestData)
	}
	return nil
}

// HandleCallback handles callback queries using middleware
func (h *HandlerWithMiddleware) HandleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	update := &tgbotapi.Update{CallbackQuery: callback}
	requestData := middleware.NewRequestDataFromUpdate(update)
	return h.callbackHandler(ctx, requestData)
}

// handleMessageWithMiddleware is the actual message handler used by middleware
func (h *HandlerWithMiddleware) handleMessageWithMiddleware(ctx context.Context, data interface{}) error {
	requestData, ok := data.(*middleware.RequestData)
	if !ok {
		return fmt.Errorf("invalid request data type")
	}

	if requestData.Message == nil {
		return fmt.Errorf("no message in request data")
	}

	message := requestData.Message

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

// handleCallbackWithMiddleware is the actual callback handler used by middleware
func (h *HandlerWithMiddleware) handleCallbackWithMiddleware(ctx context.Context, data interface{}) error {
	requestData, ok := data.(*middleware.RequestData)
	if !ok {
		return fmt.Errorf("invalid request data type")
	}

	if requestData.Callback == nil {
		return fmt.Errorf("no callback in request data")
	}

	callback := requestData.Callback

	switch callback.Data {
	case "trial":
		return h.handleTrialCallback(ctx, callback)
	case "account":
		return h.handleAccountCallback(ctx, callback)
	case "help":
		return h.handleHelpCallback(ctx, callback)
	default:
		return h.handleUnknownCallback(ctx, callback)
	}
}

// Individual handler methods (reuse existing logic from the original handler)
func (h *HandlerWithMiddleware) handleStart(ctx context.Context, message *tgbotapi.Message) error {
	user, err := h.userService.RegisterUser(
		ctx,
		message.From.ID,
		message.From.UserName,
		message.From.FirstName,
		message.From.LastName,
	)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	welcomeText := fmt.Sprintf(
		"🎉 Welcome to Arcanus VPN, %s!\n\n"+
			"🔐 Secure, private, and fast VPN service\n"+
			"📊 You have %.1f MB of free trial data\n\n"+
			"Choose an option below:",
		user.FirstName,
		float64(user.QuotaLimit)/(1024*1024),
	)

	keyboard := utils.CreateMainKeyboard()
	return h.sendMessage(message.Chat.ID, welcomeText, keyboard)
}

func (h *HandlerWithMiddleware) handleAccount(ctx context.Context, message *tgbotapi.Message) error {
	user, err := h.userService.GetUser(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	quotaUsedMB := float64(user.QuotaUsed) / (1024 * 1024)
	quotaLimitMB := float64(user.QuotaLimit) / (1024 * 1024)
	quotaUsagePercentage := user.GetQuotaUsagePercentage()

	accountText := fmt.Sprintf(
		"👤 **Your Account**\n\n"+
			"📊 **Usage Statistics:**\n"+
			"• Used: %.2f MB / %.1f MB\n"+
			"• Progress: %.1f%%\n"+
			"• Status: %s\n\n"+
			"📅 Member since: %s",
		quotaUsedMB,
		quotaLimitMB,
		quotaUsagePercentage,
		user.Status,
		user.CreatedAt.Format("January 2, 2006"),
	)

	keyboard := utils.CreateAccountKeyboard()
	return h.sendMessage(message.Chat.ID, accountText, keyboard)
}

func (h *HandlerWithMiddleware) handleHelp(ctx context.Context, message *tgbotapi.Message) error {
	helpText := "🤖 **Arcanus VPN Bot Help**\n\n" +
		"**Commands:**\n" +
		"• /start - Register and get started\n" +
		"• /account - View your account details\n" +
		"• /help - Show this help message\n\n" +
		"**Features:**\n" +
		"• 🔐 Secure VPN connection\n" +
		"• 📊 50MB free trial\n" +
		"• ⚡ Fast and reliable\n" +
		"• 🛡️ Privacy-focused\n\n" +
		"**Support:**\n" +
		"For technical support, contact @support"

	keyboard := utils.CreateHelpKeyboard()
	return h.sendMessage(message.Chat.ID, helpText, keyboard)
}

func (h *HandlerWithMiddleware) handleUnknownCommand(ctx context.Context, message *tgbotapi.Message) error {
	unknownText := "❓ Unknown command. Use /help to see available commands."
	keyboard := utils.CreateMainKeyboard()
	return h.sendMessage(message.Chat.ID, unknownText, keyboard)
}

// Callback handlers
func (h *HandlerWithMiddleware) handleTrialCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	if err := h.answerCallback(callback.ID, "🎉 Activating your free trial..."); err != nil {
		return err
	}

	if err := h.userService.ActivateTrial(ctx, callback.From.ID); err != nil {
		return fmt.Errorf("failed to activate trial: %w", err)
	}

	user, err := h.userService.GetUser(ctx, callback.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	successText := fmt.Sprintf(
		"🎉 **Free Trial Activated!**\n\n"+
			"✅ You now have %.1f MB of free VPN data\n"+
			"🔐 Your connection is secure and private\n"+
			"⚡ Enjoy fast, unlimited browsing!\n\n"+
			"Use /account to track your usage.",
		float64(user.QuotaLimit)/(1024*1024),
	)

	keyboard := utils.CreateTrialKeyboard()
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, successText, keyboard)
}

func (h *HandlerWithMiddleware) handleAccountCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	if err := h.answerCallback(callback.ID, "📊 Loading account details..."); err != nil {
		return err
	}

	user, err := h.userService.GetUser(ctx, callback.From.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	quotaUsedMB := float64(user.QuotaUsed) / (1024 * 1024)
	quotaLimitMB := float64(user.QuotaLimit) / (1024 * 1024)

	accountText := fmt.Sprintf(
		"👤 **Account Details**\n\n"+
			"📊 **Usage:**\n"+
			"• Used: %.2f MB / %.1f MB\n"+
			"• Remaining: %.2f MB\n"+
			"• Status: %s\n\n"+
			"📅 Joined: %s",
		quotaUsedMB,
		quotaLimitMB,
		quotaLimitMB-quotaUsedMB,
		user.Status,
		user.CreatedAt.Format("Jan 2, 2006"),
	)

	keyboard := utils.CreateAccountKeyboard()
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, accountText, keyboard)
}

func (h *HandlerWithMiddleware) handleHelpCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	if err := h.answerCallback(callback.ID, "❓ Loading help..."); err != nil {
		return err
	}

	helpText := "🤖 **Arcanus VPN Bot Help**\n\n" +
		"**Commands:**\n" +
		"• /start - Register and get started\n" +
		"• /account - View your account details\n" +
		"• /help - Show this help message\n\n" +
		"**Features:**\n" +
		"• 🔐 Secure VPN connection\n" +
		"• 📊 50MB free trial\n" +
		"• ⚡ Fast and reliable\n" +
		"• 🛡️ Privacy-focused\n\n" +
		"**Support:**\n" +
		"For technical support, contact @support"

	keyboard := utils.CreateBackKeyboard("main")
	return h.editMessage(callback.Message.Chat.ID, callback.Message.MessageID, helpText, keyboard)
}

func (h *HandlerWithMiddleware) handleUnknownCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	return h.answerCallback(callback.ID, "❓ Unknown action. Please try again.")
}

// Helper methods (reuse from original handler)
func (h *HandlerWithMiddleware) sendMessage(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	sentMessage, err := h.botAPI.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"chat_id": chatID,
		"message_id": sentMessage.MessageID,
		"text": text,
	}).Info("Message sent successfully")

	return nil
}

func (h *HandlerWithMiddleware) editMessage(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard

	_, err := h.botAPI.Send(edit)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"chat_id": chatID,
		"message_id": messageID,
		"text": text,
	}).Info("Message edited successfully")

	return nil
}

func (h *HandlerWithMiddleware) answerCallback(callbackID, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := h.botAPI.Request(callback)
	return err
}
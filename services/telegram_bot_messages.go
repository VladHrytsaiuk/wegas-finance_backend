package services

import (
	"fmt"
	"html"
	"strings"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
)

func buildTelegramWelcomeMessage() string {
	return strings.Join([]string{
		"<b>Telegram підключено</b>",
		"",
		"Надсилайте сюди:",
		"• XML-файл чека",
		"• посилання на електронний чек",
		"",
		"Правила:",
		"1. Один файл або одне посилання за повідомлення.",
		"2. Після обробки бот відповість, чи чек збережено.",
		"3. Якщо парсинг не вдався, бот одразу напише про помилку.",
		"",
		"Фото чеків поки що не підтримуються.",
	}, "\n")
}

func buildTelegramHelpMessage() string {
	return strings.Join([]string{
		"<b>Що можна надсилати</b>",
		"",
		"• XML-файл чека",
		"• посилання на електронний чек",
		"",
		"Бот відповість:",
		"• чи чек збережено успішно",
		"• або що саме не вдалося обробити",
	}, "\n")
}

func buildTelegramParseErrorMessage(kind string) string {
	target := "чек"
	if kind == "url" {
		target = "посилання на чек"
	}
	return fmt.Sprintf(
		"<b>Не вдалося обробити %s</b>\n\nПеревірте, що це підтримуваний XML або коректне посилання на електронний чек.",
		target,
	)
}

func buildTelegramSavedMessage(entry *models.InboxEntry) string {
	if entry == nil {
		return "<b>Чек оброблено</b>"
	}

	amount := formatAmount(entry.Total, entry.Currency)
	merchant := strings.TrimSpace(entry.Merchant)
	if merchant == "" && strings.TrimSpace(entry.Note) != "" {
		merchant = strings.TrimSpace(entry.Note)
	}
	itemCount := 0
	if len(entry.ReceiptSource.Items) > 0 {
		itemCount = len(entry.ReceiptSource.Items)
	}

	lines := []string{
		"<b>Чек збережено</b>",
		"",
		fmt.Sprintf("Сума: <b>%s</b>", amount),
	}
	if merchant != "" {
		lines = append(lines, fmt.Sprintf("Магазин: %s", html.EscapeString(merchant)))
	}
	if itemCount > 0 {
		lines = append(lines, fmt.Sprintf("Позицій: %d", itemCount))
	}
	lines = append(lines, "", "Завершіть обробку чека у застосунку.")
	return strings.Join(lines, "\n")
}

package utils

// ParseCardDetails приймає масив PAN-ів і повертає (last4, paymentSystem)
func ParseCardDetails(maskedPans []string) (string, string) {
	if len(maskedPans) == 0 {
		return "", "mastercard" // Дефолт, якщо пусто
	}

	// Зазвичай беремо перший PAN з масиву (це основна картка)
	pan := maskedPans[0] // Наприклад: "537541******1234"

	// 1. Витягуємо останні 4 цифри
	last4 := "0000"
	if len(pan) >= 4 {
		last4 = pan[len(pan)-4:]
	}

	// 2. Визначаємо платіжну систему за першою цифрою
	// Visa: починається на 4
	// Mastercard: починається на 5 або 2 (нові серії)
	paymentSystem := "mastercard" // Дефолт

	if len(pan) > 0 {
		firstDigit := pan[0]
		if firstDigit == '4' {
			paymentSystem = "visa"
		} else if firstDigit == '5' || firstDigit == '2' {
			paymentSystem = "mastercard"
		}
	}

	return last4, paymentSystem
}
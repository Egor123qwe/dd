package balance

import "context"

type Balance interface {
	// SettleRent списывает cost с клиента и начисляет cost*merchantRate продавцу в одной транзакции.
	SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost, merchantRate float64) error
}

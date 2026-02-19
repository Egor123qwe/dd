package port

import (
	"context"
)

type BalanceRepo interface {
	GetBalance(ctx context.Context, userID int) (float64, error)
	TopUp(ctx context.Context, userID int, amount float64) error
	Withdraw(ctx context.Context, userID int, amount float64) error
	// SettleRent списывает cost с клиента и начисляет cost*merchantRate продавцу в одной транзакции.
	SettleRent(ctx context.Context, clientUserID, merchantUserID int, cost float64, merchantRate float64) error
}

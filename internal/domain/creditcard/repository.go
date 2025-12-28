package creditcard

import (
	"context"

	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Repository interface {
	CreateCreditCard(ctx context.Context, card *CreditCard) error
	UpdateCreditCard(ctx context.Context, card *CreditCard) error
	DeleteCreditCard(ctx context.Context, cardID, userID ulid.ULID) error
	GetCreditCardById(ctx context.Context, cardID, userID ulid.ULID) (*CreditCard, error)
	GetCreditCardsByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*CreditCard, int64, error)
	GetCreditCardByAccountId(ctx context.Context, accountID, userID ulid.ULID) (*CreditCard, error)
	UpdateAvailableLimit(ctx context.Context, cardID ulid.ULID, amount float64) error

	CreateInvoice(ctx context.Context, invoice *Invoice) error
	UpdateInvoice(ctx context.Context, invoice *Invoice) error
	GetInvoiceById(ctx context.Context, invoiceID, userID ulid.ULID) (*Invoice, error)
	GetInvoicesByCreditCardId(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Invoice, int64, error)
	GetCurrentInvoice(ctx context.Context, cardID, userID ulid.ULID) (*Invoice, error)
	GetInvoiceByReference(ctx context.Context, cardID ulid.ULID, month, year int) (*Invoice, error)

	CreateTransaction(ctx context.Context, transaction *CreditCardTransaction) error
	GetTransactionsByInvoice(ctx context.Context, invoiceID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*CreditCardTransaction, int64, error)
	GetTransactionsByCreditCard(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*CreditCardTransaction, int64, error)
}

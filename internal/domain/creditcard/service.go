package creditcard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Fynance/internal/domain/account"
	"Fynance/internal/domain/user"
	appErrors "Fynance/internal/errors"
	"Fynance/internal/pkg"

	"github.com/oklog/ulid/v2"
)

type Service struct {
	Repository     Repository
	AccountService *account.Service
	UserService    *user.Service
}

func (s *Service) CreateCreditCard(ctx context.Context, req *CreateCreditCardRequest) (*CreditCard, error) {
	if err := s.ensureUserExists(ctx, req.UserId); err != nil {
		return nil, err
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, req.AccountId, req.UserId)
	if err != nil {
		return nil, err
	}

	if accountEntity.Type == account.TypeCreditCard {
		return nil, appErrors.NewValidationError("account_id", "conta nao pode ser do tipo CREDIT_CARD")
	}

	existingCard, _ := s.Repository.GetCreditCardByAccountId(ctx, req.AccountId, req.UserId)
	if existingCard != nil {
		return nil, appErrors.NewValidationError("account_id", "conta ja possui cartao de credito vinculado")
	}

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now()
	card := &CreditCard{
		Id:             pkg.GenerateULIDObject(),
		UserId:         req.UserId,
		AccountId:      req.AccountId,
		Name:           strings.TrimSpace(req.Name),
		CreditLimit:    req.CreditLimit,
		AvailableLimit: req.CreditLimit,
		ClosingDay:     req.ClosingDay,
		DueDay:         req.DueDay,
		Brand:          req.Brand,
		LastFourDigits: req.LastFourDigits,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.Repository.CreateCreditCard(ctx, card); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	creditCardAccount := &account.CreateAccountRequest{
		UserId:         req.UserId,
		Name:           card.Name,
		Type:           account.TypeCreditCard,
		InitialBalance: 0,
		Color:          "",
		Icon:           "credit-card",
		IncludeInTotal: false,
		CreditCardId:   &card.Id,
	}

	_, err = s.AccountService.CreateAccount(ctx, creditCardAccount)
	if err != nil {
		_ = s.Repository.DeleteCreditCard(ctx, card.Id, req.UserId)
		return nil, appErrors.NewDatabaseError(fmt.Errorf("erro ao criar conta do cartao: %w", err))
	}

	return card, nil
}

func (s *Service) UpdateCreditCard(ctx context.Context, cardID, userID ulid.ULID, req *UpdateCreditCardRequest) error {
	card, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return appErrors.NewValidationError("name", "nao pode ser vazio")
		}
		card.Name = name
	}

	if req.CreditLimit != nil {
		if *req.CreditLimit < 0 {
			return appErrors.NewValidationError("credit_limit", "deve ser maior ou igual a zero")
		}
		limitDiff := *req.CreditLimit - card.CreditLimit
		card.CreditLimit = *req.CreditLimit
		card.AvailableLimit += limitDiff
		if card.AvailableLimit < 0 {
			return appErrors.NewValidationError("credit_limit", "limite disponivel nao pode ser negativo")
		}
	}

	if req.ClosingDay != nil {
		if *req.ClosingDay < 1 || *req.ClosingDay > 31 {
			return appErrors.NewValidationError("closing_day", "deve estar entre 1 e 31")
		}
		card.ClosingDay = *req.ClosingDay
	}

	if req.DueDay != nil {
		if *req.DueDay < 1 || *req.DueDay > 31 {
			return appErrors.NewValidationError("due_day", "deve estar entre 1 e 31")
		}
		card.DueDay = *req.DueDay
	}

	if req.Brand != nil {
		if !req.Brand.IsValid() {
			return appErrors.NewValidationError("brand", "bandeira invalida")
		}
		card.Brand = *req.Brand
	}

	if req.LastFourDigits != nil {
		card.LastFourDigits = *req.LastFourDigits
	}

	if req.IsActive != nil {
		card.IsActive = *req.IsActive
	}

	card.UpdatedAt = time.Now()

	return s.Repository.UpdateCreditCard(ctx, card)
}

func (s *Service) DeleteCreditCard(ctx context.Context, cardID, userID ulid.ULID) error {
	_, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return err
	}

	currentInvoice, err := s.Repository.GetCurrentInvoice(ctx, cardID, userID)
	if err == nil && currentInvoice != nil && currentInvoice.TotalAmount > 0 {
		return appErrors.NewValidationError("credit_card", "Cartão possui fatura em aberto, não pode remover")
	}

	creditCardAccount, err := s.AccountService.Repository.GetByCreditCardId(ctx, cardID, userID)
	if err == nil && creditCardAccount != nil {
		if err := s.AccountService.Repository.Delete(ctx, creditCardAccount.Id, userID); err != nil {
			return appErrors.NewDatabaseError(fmt.Errorf("erro ao remover conta do cartao: %w", err))
		}
	}

	return s.Repository.DeleteCreditCard(ctx, cardID, userID)
}

func (s *Service) GetCreditCardById(ctx context.Context, cardID, userID ulid.ULID) (*CreditCard, error) {
	card, err := s.Repository.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if card.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return card, nil
}

func (s *Service) ListCreditCards(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*CreditCard, int64, error) {
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return nil, 0, err
	}

	return s.Repository.GetCreditCardsByUserId(ctx, userID, pagination)
}

func (s *Service) CreateTransaction(ctx context.Context, req *CreateTransactionRequest) error {
	if err := s.ensureUserExists(ctx, req.UserId); err != nil {
		return err
	}

	card, err := s.GetCreditCardById(ctx, req.CreditCardId, req.UserId)
	if err != nil {
		return err
	}

	if !card.IsActive {
		return appErrors.NewValidationError("credit_card", "cartao nao esta ativo")
	}

	amount := req.Amount
	if amount < 0 {
		amount = -amount
	}
	if amount <= 0 {
		return appErrors.NewValidationError("amount", "valor deve ser maior que zero")
	}

	if card.AvailableLimit < amount {
		return appErrors.NewValidationError("amount", "Limite disponível insuficiente")
	}

	invoice, err := s.getOrCreateCurrentInvoice(ctx, card)
	if err != nil {
		return err
	}

	now := time.Now()
	transaction := &CreditCardTransaction{
		Id:                 pkg.GenerateULIDObject(),
		CreditCardId:       req.CreditCardId,
		InvoiceId:          invoice.Id,
		UserId:             req.UserId,
		CategoryId:         req.CategoryId,
		Amount:             amount,
		Description:        strings.TrimSpace(req.Description),
		Date:               req.Date,
		Installments:       req.Installments,
		CurrentInstallment: 1,
		IsRecurring:        req.IsRecurring,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.Repository.CreateTransaction(ctx, transaction); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	invoice.TotalAmount += amount
	if err := s.Repository.UpdateInvoice(ctx, invoice); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	deductionAmount := -amount
	if err := s.Repository.UpdateAvailableLimit(ctx, req.CreditCardId, deductionAmount); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) PayInvoice(ctx context.Context, cardID, invoiceID, accountID, userID ulid.ULID, amount float64) error {
	if amount <= 0 {
		return appErrors.NewValidationError("amount", "deve ser maior que zero")
	}

	_, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return err
	}

	invoice, err := s.Repository.GetInvoiceById(ctx, invoiceID, userID)
	if err != nil {
		return appErrors.ErrNotFound.WithError(err)
	}

	if invoice.CreditCardId != cardID {
		return appErrors.NewValidationError("invoice_id", "fatura nao pertence a este cartao")
	}

	if invoice.Status == InvoicePaid {
		return appErrors.NewValidationError("invoice", "fatura ja esta paga")
	}

	accountEntity, err := s.AccountService.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return err
	}

	if accountEntity.Type == account.TypeCreditCard {
		return appErrors.NewValidationError("account_id", "conta de pagamento nao pode ser cartao de credito")
	}

	if accountEntity.Balance < amount {
		return appErrors.NewValidationError("amount", "saldo insuficiente")
	}

	remainingAmount := invoice.TotalAmount - invoice.PaidAmount
	if amount > remainingAmount {
		amount = remainingAmount
	}

	if err := s.AccountService.UpdateBalance(ctx, accountID, userID, -amount); err != nil {
		return err
	}

	invoice.PaidAmount += amount
	now := time.Now()

	if invoice.PaidAmount >= invoice.TotalAmount {
		invoice.Status = InvoicePaid
		invoice.PaidAt = &now
	} else {
		invoice.Status = InvoicePartial
	}

	invoice.UpdatedAt = now

	if err := s.Repository.UpdateInvoice(ctx, invoice); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	if err := s.Repository.UpdateAvailableLimit(ctx, cardID, amount); err != nil {
		return appErrors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) GetCurrentInvoice(ctx context.Context, cardID, userID ulid.ULID) (*Invoice, error) {
	card, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	invoice, err := s.Repository.GetCurrentInvoice(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	if invoice == nil {
		return s.getOrCreateCurrentInvoice(ctx, card)
	}

	return invoice, nil
}

func (s *Service) ListInvoices(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*Invoice, int64, error) {
	_, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return nil, 0, err
	}

	return s.Repository.GetInvoicesByCreditCardId(ctx, cardID, userID, pagination)
}

func (s *Service) GetInvoiceById(ctx context.Context, invoiceID, userID ulid.ULID) (*Invoice, error) {
	invoice, err := s.Repository.GetInvoiceById(ctx, invoiceID, userID)
	if err != nil {
		return nil, appErrors.ErrNotFound.WithError(err)
	}

	if invoice.UserId != userID {
		return nil, appErrors.ErrResourceNotOwned
	}

	return invoice, nil
}

func (s *Service) ListTransactions(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*CreditCardTransaction, int64, error) {
	_, err := s.GetCreditCardById(ctx, cardID, userID)
	if err != nil {
		return nil, 0, err
	}

	return s.Repository.GetTransactionsByCreditCard(ctx, cardID, userID, pagination)
}

func (s *Service) getOrCreateCurrentInvoice(ctx context.Context, card *CreditCard) (*Invoice, error) {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	invoice, err := s.Repository.GetInvoiceByReference(ctx, card.Id, currentMonth, currentYear)
	if err == nil && invoice != nil {
		return invoice, nil
	}

	closingDate := s.calculateClosingDate(currentYear, currentMonth, card.ClosingDay)
	dueDate := s.calculateDueDate(closingDate, card.DueDay)

	newInvoice := &Invoice{
		Id:             pkg.GenerateULIDObject(),
		CreditCardId:   card.Id,
		UserId:         card.UserId,
		ReferenceMonth: currentMonth,
		ReferenceYear:  currentYear,
		OpeningDate:    now,
		ClosingDate:    closingDate,
		DueDate:        dueDate,
		TotalAmount:    0,
		PaidAmount:     0,
		Status:         InvoiceOpen,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.Repository.CreateInvoice(ctx, newInvoice); err != nil {
		return nil, appErrors.NewDatabaseError(err)
	}

	return newInvoice, nil
}

func (s *Service) calculateClosingDate(year, month, day int) time.Time {
	closingDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if closingDate.Before(time.Now()) {
		closingDate = closingDate.AddDate(0, 1, 0)
	}
	return closingDate
}

func (s *Service) calculateDueDate(closingDate time.Time, dueDay int) time.Time {
	dueDate := time.Date(closingDate.Year(), closingDate.Month(), dueDay, 0, 0, 0, 0, time.UTC)
	if dueDate.Before(closingDate) {
		dueDate = dueDate.AddDate(0, 1, 0)
	}
	return dueDate
}

func (s *Service) validateCreateRequest(req *CreateCreditCardRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return appErrors.NewValidationError("name", "e obrigatorio")
	}

	if req.CreditLimit <= 0 {
		return appErrors.NewValidationError("credit_limit", "deve ser maior que zero")
	}

	if req.ClosingDay < 1 || req.ClosingDay > 31 {
		return appErrors.NewValidationError("closing_day", "deve estar entre 1 e 31")
	}

	if req.DueDay < 1 || req.DueDay > 31 {
		return appErrors.NewValidationError("due_day", "deve estar entre 1 e 31")
	}

	if !req.Brand.IsValid() {
		return appErrors.NewValidationError("brand", "bandeira invalida")
	}

	return nil
}

func (s *Service) ensureUserExists(ctx context.Context, userID ulid.ULID) error {
	if s.UserService == nil {
		return appErrors.ErrInternalServer.WithError(fmt.Errorf("servico de usuario nao configurado"))
	}
	_, err := s.UserService.GetByID(ctx, userID)
	if err != nil {
		return appErrors.ErrUserNotFound.WithError(err)
	}
	return nil
}

type CreateCreditCardRequest struct {
	UserId         ulid.ULID
	AccountId      ulid.ULID
	Name           string
	CreditLimit    float64
	ClosingDay     int
	DueDay         int
	Brand          CardBrand
	LastFourDigits string
}

type UpdateCreditCardRequest struct {
	Name           *string
	CreditLimit    *float64
	ClosingDay     *int
	DueDay         *int
	Brand          *CardBrand
	LastFourDigits *string
	IsActive       *bool
}

type CreateTransactionRequest struct {
	CreditCardId ulid.ULID
	UserId       ulid.ULID
	CategoryId   ulid.ULID
	Amount       float64
	Description  string
	Date         time.Time
	Installments int
	IsRecurring  bool
}

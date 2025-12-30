package infrastructure

import (
	"context"
	"time"

	"Fynance/internal/domain/creditcard"
	"Fynance/internal/pkg"

	"errors"

	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

type CreditCardRepository struct {
	DB *gorm.DB
}

type creditCardDB struct {
	Id             string    `gorm:"type:varchar(26);primaryKey"`
	UserId         string    `gorm:"type:varchar(26);index;not null"`
	AccountId      string    `gorm:"type:varchar(26);index;not null"`
	Name           string    `gorm:"type:varchar(100);not null"`
	CreditLimit    float64   `gorm:"type:decimal(15,2);not null"`
	AvailableLimit float64   `gorm:"type:decimal(15,2);not null"`
	ClosingDay     int       `gorm:"not null"`
	DueDay         int       `gorm:"not null"`
	Brand          string    `gorm:"type:varchar(20);not null"`
	LastFourDigits string    `gorm:"type:varchar(4)"`
	IsActive       bool      `gorm:"not null;default:true"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

func (creditCardDB) TableName() string {
	return "credit_cards"
}

type invoiceDB struct {
	Id             string     `gorm:"type:varchar(26);primaryKey"`
	CreditCardId   string     `gorm:"type:varchar(26);index;not null"`
	UserId         string     `gorm:"type:varchar(26);index;not null"`
	ReferenceMonth int        `gorm:"not null"`
	ReferenceYear  int        `gorm:"not null"`
	OpeningDate    time.Time  `gorm:"type:date;not null"`
	ClosingDate    time.Time  `gorm:"type:date;not null"`
	DueDate        time.Time  `gorm:"type:date;not null"`
	TotalAmount    float64    `gorm:"type:decimal(15,2);not null;default:0"`
	PaidAmount     float64    `gorm:"type:decimal(15,2);not null;default:0"`
	Status         string     `gorm:"type:varchar(20);not null;default:'OPEN'"`
	PaidAt         *time.Time `gorm:"type:timestamp"`
	CreatedAt      time.Time  `gorm:"not null"`
	UpdatedAt      time.Time  `gorm:"not null"`
}

func (invoiceDB) TableName() string {
	return "invoices"
}

type creditCardTransactionDB struct {
	Id                 string    `gorm:"type:varchar(26);primaryKey;column:id"`
	CreditCardId       string    `gorm:"type:varchar(26);index;not null;column:credit_card_id"`
	InvoiceId          string    `gorm:"type:varchar(26);index;not null;column:invoice_id"`
	UserId             string    `gorm:"type:varchar(26);index;not null;column:user_id"`
	CategoryId         string    `gorm:"type:varchar(26);index;not null;column:category_id"`
	CategoryName       string    `gorm:"->;column:category_name"`
	Amount             float64   `gorm:"type:decimal(15,2);not null;column:amount"`
	Description        string    `gorm:"type:varchar(255);column:description"`
	Date               time.Time `gorm:"type:date;not null;column:date"`
	Installments       int       `gorm:"not null;default:1;column:installments"`
	CurrentInstallment int       `gorm:"not null;default:1;column:current_installment"`
	IsRecurring        bool      `gorm:"not null;default:false;column:is_recurring"`
	CreatedAt          time.Time `gorm:"not null;column:created_at"`
	UpdatedAt          time.Time `gorm:"not null;column:updated_at"`
}

func (creditCardTransactionDB) TableName() string {
	return "credit_card_transactions"
}

func toDomainCreditCard(ccdb *creditCardDB) (*creditcard.CreditCard, error) {
	id, err := pkg.ParseULID(ccdb.Id)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(ccdb.UserId)
	if err != nil {
		return nil, err
	}
	aid, err := pkg.ParseULID(ccdb.AccountId)
	if err != nil {
		return nil, err
	}

	return &creditcard.CreditCard{
		Id:             id,
		UserId:         uid,
		AccountId:      aid,
		Name:           ccdb.Name,
		CreditLimit:    ccdb.CreditLimit,
		AvailableLimit: ccdb.AvailableLimit,
		ClosingDay:     ccdb.ClosingDay,
		DueDay:         ccdb.DueDay,
		Brand:          creditcard.CardBrand(ccdb.Brand),
		LastFourDigits: ccdb.LastFourDigits,
		IsActive:       ccdb.IsActive,
		CreatedAt:      ccdb.CreatedAt,
		UpdatedAt:      ccdb.UpdatedAt,
	}, nil
}

func toDBCreditCard(cc *creditcard.CreditCard) *creditCardDB {
	return &creditCardDB{
		Id:             cc.Id.String(),
		UserId:         cc.UserId.String(),
		AccountId:      cc.AccountId.String(),
		Name:           cc.Name,
		CreditLimit:    cc.CreditLimit,
		AvailableLimit: cc.AvailableLimit,
		ClosingDay:     cc.ClosingDay,
		DueDay:         cc.DueDay,
		Brand:          string(cc.Brand),
		LastFourDigits: cc.LastFourDigits,
		IsActive:       cc.IsActive,
		CreatedAt:      cc.CreatedAt,
		UpdatedAt:      cc.UpdatedAt,
	}
}

func toDomainInvoice(idb *invoiceDB) (*creditcard.Invoice, error) {
	id, err := pkg.ParseULID(idb.Id)
	if err != nil {
		return nil, err
	}
	ccid, err := pkg.ParseULID(idb.CreditCardId)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(idb.UserId)
	if err != nil {
		return nil, err
	}

	return &creditcard.Invoice{
		Id:             id,
		CreditCardId:   ccid,
		UserId:         uid,
		ReferenceMonth: idb.ReferenceMonth,
		ReferenceYear:  idb.ReferenceYear,
		OpeningDate:    idb.OpeningDate,
		ClosingDate:    idb.ClosingDate,
		DueDate:        idb.DueDate,
		TotalAmount:    idb.TotalAmount,
		PaidAmount:     idb.PaidAmount,
		Status:         creditcard.InvoiceStatus(idb.Status),
		PaidAt:         idb.PaidAt,
		CreatedAt:      idb.CreatedAt,
		UpdatedAt:      idb.UpdatedAt,
	}, nil
}

func toDBInvoice(inv *creditcard.Invoice) *invoiceDB {
	return &invoiceDB{
		Id:             inv.Id.String(),
		CreditCardId:   inv.CreditCardId.String(),
		UserId:         inv.UserId.String(),
		ReferenceMonth: inv.ReferenceMonth,
		ReferenceYear:  inv.ReferenceYear,
		OpeningDate:    inv.OpeningDate,
		ClosingDate:    inv.ClosingDate,
		DueDate:        inv.DueDate,
		TotalAmount:    inv.TotalAmount,
		PaidAmount:     inv.PaidAmount,
		Status:         string(inv.Status),
		PaidAt:         inv.PaidAt,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
	}
}

func toDomainCreditCardTransaction(tdb *creditCardTransactionDB) (*creditcard.CreditCardTransaction, error) {
	id, err := pkg.ParseULID(tdb.Id)
	if err != nil {
		return nil, err
	}
	ccid, err := pkg.ParseULID(tdb.CreditCardId)
	if err != nil {
		return nil, err
	}
	invid, err := pkg.ParseULID(tdb.InvoiceId)
	if err != nil {
		return nil, err
	}
	uid, err := pkg.ParseULID(tdb.UserId)
	if err != nil {
		return nil, err
	}
	cid, err := pkg.ParseULID(tdb.CategoryId)
	if err != nil {
		return nil, err
	}

	tx := &creditcard.CreditCardTransaction{
		Id:                 id,
		CreditCardId:       ccid,
		InvoiceId:          invid,
		UserId:             uid,
		CategoryId:         cid,
		Amount:             tdb.Amount,
		Description:        tdb.Description,
		Date:               tdb.Date,
		Installments:       tdb.Installments,
		CurrentInstallment: tdb.CurrentInstallment,
		IsRecurring:        tdb.IsRecurring,
		CreatedAt:          tdb.CreatedAt,
		UpdatedAt:          tdb.UpdatedAt,
	}
	if tdb.CategoryName != "" {
		tx.CategoryName = tdb.CategoryName
	}
	return tx, nil
}

func toDBCreditCardTransaction(t *creditcard.CreditCardTransaction) *creditCardTransactionDB {
	return &creditCardTransactionDB{
		Id:                 t.Id.String(),
		CreditCardId:       t.CreditCardId.String(),
		InvoiceId:          t.InvoiceId.String(),
		UserId:             t.UserId.String(),
		CategoryId:         t.CategoryId.String(),
		Amount:             t.Amount,
		Description:        t.Description,
		Date:               t.Date,
		Installments:       t.Installments,
		CurrentInstallment: t.CurrentInstallment,
		IsRecurring:        t.IsRecurring,
		CreatedAt:          t.CreatedAt,
		UpdatedAt:          t.UpdatedAt,
	}
}

func (r *CreditCardRepository) CreateCreditCard(ctx context.Context, card *creditcard.CreditCard) error {
	ccdb := toDBCreditCard(card)
	return r.DB.WithContext(ctx).Table("credit_cards").Create(ccdb).Error
}

func (r *CreditCardRepository) UpdateCreditCard(ctx context.Context, card *creditcard.CreditCard) error {
	ccdb := toDBCreditCard(card)
	return r.DB.WithContext(ctx).Model(&creditCardDB{}).Where("id = ? AND user_id = ?", ccdb.Id, ccdb.UserId).Updates(ccdb).Error
}

func (r *CreditCardRepository) DeleteCreditCard(ctx context.Context, cardID, userID ulid.ULID) error {
	return r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", cardID.String(), userID.String()).Delete(&creditCardDB{}).Error
}

func (r *CreditCardRepository) GetCreditCardById(ctx context.Context, cardID, userID ulid.ULID) (*creditcard.CreditCard, error) {
	var ccdb creditCardDB
	err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", cardID.String(), userID.String()).First(&ccdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainCreditCard(&ccdb)
}

func (r *CreditCardRepository) GetCreditCardsByUserId(ctx context.Context, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*creditcard.CreditCard, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("credit_cards").Where("user_id = ?", userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []creditCardDB
	err := baseQuery.Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	cards := make([]*creditcard.CreditCard, 0, len(rows))
	for i := range rows {
		card, err := toDomainCreditCard(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		cards = append(cards, card)
	}
	return cards, total, nil
}

func (r *CreditCardRepository) GetCreditCardByAccountId(ctx context.Context, accountID, userID ulid.ULID) (*creditcard.CreditCard, error) {
	var ccdb creditCardDB
	err := r.DB.WithContext(ctx).Where("account_id = ? AND user_id = ?", accountID.String(), userID.String()).First(&ccdb).Error
	if err != nil {
		return nil, err
	}
	return toDomainCreditCard(&ccdb)
}

func (r *CreditCardRepository) UpdateAvailableLimit(ctx context.Context, cardID ulid.ULID, amount float64) error {
	return r.DB.WithContext(ctx).Model(&creditCardDB{}).
		Where("id = ?", cardID.String()).
		Update("available_limit", gorm.Expr("available_limit + ?", amount)).Error
}

func (r *CreditCardRepository) CreateInvoice(ctx context.Context, invoice *creditcard.Invoice) error {
	idb := toDBInvoice(invoice)
	return r.DB.WithContext(ctx).Table("invoices").Create(idb).Error
}

func (r *CreditCardRepository) UpdateInvoice(ctx context.Context, invoice *creditcard.Invoice) error {
	idb := toDBInvoice(invoice)
	return r.DB.WithContext(ctx).Model(&invoiceDB{}).Where("id = ? AND user_id = ?", idb.Id, idb.UserId).Updates(idb).Error
}

func (r *CreditCardRepository) GetInvoiceById(ctx context.Context, invoiceID, userID ulid.ULID) (*creditcard.Invoice, error) {
	var idb invoiceDB
	err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", invoiceID.String(), userID.String()).First(&idb).Error
	if err != nil {
		return nil, err
	}
	return toDomainInvoice(&idb)
}

func (r *CreditCardRepository) GetInvoicesByCreditCardId(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*creditcard.Invoice, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	baseQuery := r.DB.WithContext(ctx).Table("invoices").Where("credit_card_id = ? AND user_id = ?", cardID.String(), userID.String())

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []invoiceDB
	err := baseQuery.Order("reference_year DESC, reference_month DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	invoices := make([]*creditcard.Invoice, 0, len(rows))
	for i := range rows {
		invoice, err := toDomainInvoice(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, total, nil
}

func (r *CreditCardRepository) GetCurrentInvoice(ctx context.Context, cardID, userID ulid.ULID) (*creditcard.Invoice, error) {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	var idb invoiceDB
	err := r.DB.WithContext(ctx).
		Where("credit_card_id = ? AND user_id = ? AND reference_month = ? AND reference_year = ? AND status = ?",
			cardID.String(), userID.String(), currentMonth, currentYear, string(creditcard.InvoiceOpen)).
		First(&idb).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainInvoice(&idb)
}

func (r *CreditCardRepository) GetInvoiceByReference(ctx context.Context, cardID ulid.ULID, month, year int) (*creditcard.Invoice, error) {
	var idb invoiceDB
	err := r.DB.WithContext(ctx).
		Where("credit_card_id = ? AND reference_month = ? AND reference_year = ?",
			cardID.String(), month, year).
		First(&idb).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainInvoice(&idb)
}

func (r *CreditCardRepository) CreateTransaction(ctx context.Context, transaction *creditcard.CreditCardTransaction) error {
	tdb := toDBCreditCardTransaction(transaction)
	return r.DB.WithContext(ctx).Table("credit_card_transactions").Create(tdb).Error
}

func (r *CreditCardRepository) GetTransactionsByInvoice(ctx context.Context, invoiceID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*creditcard.CreditCardTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	countQuery := r.DB.WithContext(ctx).Table("credit_card_transactions t").Where("t.invoice_id = ? AND t.user_id = ?", invoiceID.String(), userID.String())
	dataQuery := r.DB.WithContext(ctx).Table("credit_card_transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.invoice_id = ? AND t.user_id = ?", invoiceID.String(), userID.String())

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []creditCardTransactionDB
	err := dataQuery.Order("t.date DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*creditcard.CreditCardTransaction, 0, len(rows))
	for i := range rows {
		transaction, err := toDomainCreditCardTransaction(&rows[i])
		if err != nil {
			continue
		}
		transactions = append(transactions, transaction)
	}
	return transactions, total, nil
}

func (r *CreditCardRepository) GetTransactionsByCreditCard(ctx context.Context, cardID, userID ulid.ULID, pagination *pkg.PaginationParams) ([]*creditcard.CreditCardTransaction, int64, error) {
	if pagination == nil {
		pagination = &pkg.PaginationParams{Page: 1, Limit: 10}
	}
	pagination.Normalize()

	countQuery := r.DB.WithContext(ctx).Table("credit_card_transactions t").Where("t.credit_card_id = ? AND t.user_id = ?", cardID.String(), userID.String())
	dataQuery := r.DB.WithContext(ctx).Table("credit_card_transactions t").
		Select("t.*, c.name as category_name").
		Joins("LEFT JOIN categories c ON t.category_id = c.id").
		Where("t.credit_card_id = ? AND t.user_id = ?", cardID.String(), userID.String())

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []creditCardTransactionDB
	err := dataQuery.Order("t.date DESC").
		Offset(pagination.Offset()).
		Limit(pagination.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	transactions := make([]*creditcard.CreditCardTransaction, 0, len(rows))
	for i := range rows {
		transaction, err := toDomainCreditCardTransaction(&rows[i])
		if err != nil {
			continue
		}
		transactions = append(transactions, transaction)
	}
	return transactions, total, nil
}

package account

type AccountType string

const (
	TypeChecking   AccountType = "CHECKING"
	TypeSavings    AccountType = "SAVINGS"
	TypeCreditCard AccountType = "CREDIT_CARD"
	TypeCash       AccountType = "CASH"
	TypeInvestment AccountType = "INVESTMENT"
	TypeOther      AccountType = "OTHER"
)

func (t AccountType) IsValid() bool {
	switch t {
	case TypeChecking, TypeSavings, TypeCreditCard, TypeCash, TypeInvestment, TypeOther:
		return true
	}
	return false
}

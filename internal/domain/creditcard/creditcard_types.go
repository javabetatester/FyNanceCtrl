package creditcard

type CardBrand string

const (
	BrandVisa       CardBrand = "VISA"
	BrandMastercard CardBrand = "MASTERCARD"
	BrandElo        CardBrand = "ELO"
	BrandAmex       CardBrand = "AMEX"
	BrandHipercard  CardBrand = "HIPERCARD"
	BrandOther      CardBrand = "OTHER"
)

func (b CardBrand) IsValid() bool {
	switch b {
	case BrandVisa, BrandMastercard, BrandElo, BrandAmex, BrandHipercard, BrandOther:
		return true
	}
	return false
}

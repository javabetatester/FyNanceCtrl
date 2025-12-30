package category

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// GroupType define o tipo do grupo de categoria
type GroupType string

const (
	GroupTypeExpense GroupType = "EXPENSE"
	GroupTypeReceipt GroupType = "RECEIPT"
)

// CategoryGroup representa um grupo de categorias (Essencial, Variável, Eventual, Lazer)
type CategoryGroup struct {
	Id        ulid.ULID `gorm:"type:varchar(26);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	Type      GroupType `gorm:"type:varchar(10);not null;index:idx_category_groups_type" json:"type"`
	Icon      string    `gorm:"type:varchar(50)" json:"icon"`
	Color     string    `gorm:"type:varchar(7)" json:"color"`
	SortOrder int       `gorm:"default:0;index:idx_category_groups_sort" json:"sortOrder"`
	IsSystem  bool      `gorm:"default:true" json:"isSystem"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (CategoryGroup) TableName() string {
	return "category_groups"
}

// DefaultExpenseGroup define a estrutura para grupos de despesa padrão
type DefaultExpenseGroup struct {
	Name  string
	Icon  string
	Color string
	Order int
}

// DefaultReceiptGroup define a estrutura para grupos de receita padrão
type DefaultReceiptGroup struct {
	Name  string
	Icon  string
	Color string
	Order int
}

// DefaultExpenseGroups são os grupos padrão de despesas
var DefaultExpenseGroups = []DefaultExpenseGroup{
	{Name: "Essencial", Icon: "shield", Color: "#4CAF50", Order: 1},
	{Name: "Variável", Icon: "sliders", Color: "#2196F3", Order: 2},
	{Name: "Eventual", Icon: "calendar", Color: "#FF9800", Order: 3},
	{Name: "Lazer", Icon: "gamepad-2", Color: "#9C27B0", Order: 4},
}

// DefaultReceiptGroups são os grupos padrão de receitas
var DefaultReceiptGroups = []DefaultReceiptGroup{
	{Name: "Trabalho", Icon: "briefcase", Color: "#4CAF50", Order: 1},
	{Name: "Investimentos", Icon: "trending-up", Color: "#2196F3", Order: 2},
	{Name: "Extras", Icon: "gift", Color: "#FF9800", Order: 3},
}

// DefaultCategoriesByGroup define as categorias padrão organizadas por grupo
var DefaultCategoriesByGroup = map[string]map[string][]string{
	"Essencial": {
		"Moradia":      {"Aluguel", "Condomínio", "IPTU", "Manutenção Casa"},
		"Contas Fixas": {"Energia", "Água", "Gás", "Internet", "Telefone"},
		"Alimentação":  {"Supermercado", "Feira", "Açougue", "Padaria"},
		"Transporte":   {"Combustível", "Transporte Público", "Manutenção Veículo", "Estacionamento"},
		"Saúde":        {"Plano de Saúde", "Farmácia", "Consultas", "Exames"},
	},
	"Variável": {
		"Alimentação Extra": {"Restaurantes", "Delivery", "Lanches", "Cafeterias"},
		"Vestuário":         {"Roupas", "Calçados", "Acessórios"},
		"Cuidados Pessoais": {"Beleza", "Academia", "Higiene", "Cosméticos"},
		"Educação":          {"Cursos", "Livros", "Material Escolar", "Assinaturas"},
	},
	"Eventual": {
		"_root": {"Presentes", "Manutenções", "Taxas e Impostos", "Emergências", "Reparos", "Documentos"},
	},
	"Lazer": {
		"Entretenimento": {"Streaming", "Cinema", "Shows", "Jogos", "Música"},
		"Viagens":        {"Hospedagem", "Passagens", "Passeios", "Alimentação Viagem"},
		"Hobbies":        {"Esportes", "Artesanato", "Coleções"},
		"Social":         {"Bares", "Eventos", "Festas", "Comemorações"},
	},
}

// DefaultReceiptCategoriesByGroup define as categorias de receita por grupo
var DefaultReceiptCategoriesByGroup = map[string]map[string][]string{
	"Trabalho": {
		"_root": {"Salário", "Freelance", "Bonificações", "Comissões", "13º Salário", "Férias"},
	},
	"Investimentos": {
		"_root": {"Dividendos", "Rendimentos", "Juros", "Aluguel Recebido", "Lucro Venda"},
	},
	"Extras": {
		"_root": {"Prêmios", "Presentes Recebidos", "Reembolsos", "Vendas", "Cashback"},
	},
}

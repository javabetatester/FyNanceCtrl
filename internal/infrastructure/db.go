package infrastructure

import (
	"Fynance/config"
	"Fynance/internal/domain/account"
	"Fynance/internal/domain/budget"
	"Fynance/internal/domain/creditcard"
	"Fynance/internal/domain/goal"
	"Fynance/internal/domain/investment"
	"Fynance/internal/domain/recurring"
	"Fynance/internal/domain/transaction"
	"Fynance/internal/domain/user"
	"Fynance/internal/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDb(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		logger.Error().
			Err(err).
			Str("host", cfg.Database.Host).
			Int("port", cfg.Database.Port).
			Str("database", cfg.Database.DBName).
			Msg("Falha ao conectar ao banco de dados")
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().Err(err).Msg("Falha ao obter instância do banco de dados")
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Info().
		Str("host", cfg.Database.Host).
		Int("port", cfg.Database.Port).
		Str("database", cfg.Database.DBName).
		Msg("Conexão com banco de dados estabelecida com sucesso")

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	logger.Info().Msg("Executando migrations...")

	if err := removeUniqueConstraintOnUserName(db); err != nil {
		logger.Warn().Err(err).Msg("Aviso ao remover constraint única do campo name da tabela users")
	}

	if err := fixBudgetMonthYearTypes(db); err != nil {
		logger.Warn().Err(err).Msg("Aviso ao corrigir tipos das colunas month e year da tabela budgets")
	}

	entities := []interface{}{
		&user.User{},
		&goal.Goal{},
		&goal.Contribution{},
		&transaction.Transaction{},
		&transaction.Category{},
		&investment.Investment{},
		&account.Account{},
		&budget.Budget{},
		&recurring.RecurringTransaction{},
		&creditcard.CreditCard{},
		&creditcard.Invoice{},
		&creditcard.CreditCardTransaction{},
	}

	for _, entity := range entities {
		if err := db.AutoMigrate(entity); err != nil {
			logger.Error().
				Err(err).
				Str("entity", getEntityName(entity)).
				Msg("Erro ao migrar entidade")
			return err
		}
	}

	logger.Info().Msg("Migrations executadas com sucesso!")
	return nil
}

func removeUniqueConstraintOnUserName(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	query := `
		SELECT constraint_name
		FROM information_schema.table_constraints
		WHERE table_name = 'users'
		AND constraint_type = 'UNIQUE'
		AND constraint_name != 'idx_users_email'
		AND constraint_name IN (
			SELECT constraint_name
			FROM information_schema.key_column_usage
			WHERE table_name = 'users'
			AND column_name = 'name'
		)
	`

	rows, err := sqlDB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var constraints []string
	for rows.Next() {
		var constraintName string
		if err := rows.Scan(&constraintName); err != nil {
			continue
		}
		constraints = append(constraints, constraintName)
	}

	for _, constraintName := range constraints {
		dropQuery := `ALTER TABLE users DROP CONSTRAINT IF EXISTS ` + constraintName
		if _, err := sqlDB.Exec(dropQuery); err != nil {
			logger.Warn().
				Err(err).
				Str("constraint", constraintName).
				Msg("Não foi possível remover constraint única do campo name")
			continue
		}
		logger.Info().
			Str("constraint", constraintName).
			Msg("Constraint única removida do campo name da tabela users")
	}

	return nil
}

func fixBudgetMonthYearTypes(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	checkQuery := `
		SELECT data_type 
		FROM information_schema.columns 
		WHERE table_name = 'budgets' 
		AND column_name = 'year'
	`

	var dataType string
	err = sqlDB.QueryRow(checkQuery).Scan(&dataType)
	if err != nil {
		return nil
	}

	if dataType == "numeric" || dataType == "decimal" {
		logger.Info().Msg("Corrigindo tipos das colunas month e year na tabela budgets...")

		queries := []string{
			`ALTER TABLE budgets ALTER COLUMN month TYPE integer USING month::integer`,
			`ALTER TABLE budgets ALTER COLUMN year TYPE integer USING year::integer`,
		}

		for _, query := range queries {
			if _, err := sqlDB.Exec(query); err != nil {
				logger.Warn().
					Err(err).
					Str("query", query).
					Msg("Erro ao executar migração de tipos")
				return err
			}
		}

		logger.Info().Msg("Tipos das colunas month e year corrigidos com sucesso!")
	}

	return nil
}

func getEntityName(entity interface{}) string {
	switch entity.(type) {
	case *user.User:
		return "User"
	case *goal.Goal:
		return "Goal"
	case *goal.Contribution:
		return "GoalContribution"
	case *transaction.Transaction:
		return "Transaction"
	case *transaction.Category:
		return "Category"
	case *investment.Investment:
		return "Investment"
	case *account.Account:
		return "Account"
	case *budget.Budget:
		return "Budget"
	case *recurring.RecurringTransaction:
		return "RecurringTransaction"
	case *creditcard.CreditCard:
		return "CreditCard"
	case *creditcard.Invoice:
		return "Invoice"
	case *creditcard.CreditCardTransaction:
		return "CreditCardTransaction"
	default:
		return "Unknown"
	}
}

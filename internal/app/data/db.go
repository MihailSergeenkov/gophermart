package data

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type DBStorage struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDBStorage(ctx context.Context, logger *zap.Logger, dbURI string) (*DBStorage, error) {
	if err := runMigrations(dbURI); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	pool, err := initPool(ctx, logger, dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	s := &DBStorage{
		logger: logger,
		pool:   pool,
	}

	return s, nil
}

func (s *DBStorage) GetUserById(ctx context.Context, userID int) (models.User, error) {
	const query = `SELECT id, login, password FROM users WHERE id = $1 LIMIT 1`

	row := s.pool.QueryRow(ctx, query, userID)

	var u models.User
	err := row.Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%w with ID: %d", ErrUserNotFound, userID)
		}

		return models.User{}, fmt.Errorf("failed to scan a response row: %w", err)
	}

	return u, nil
}

func (s *DBStorage) GetUserByLogin(ctx context.Context, userLogin string) (models.User, error) {
	const query = `SELECT id, login, password FROM users WHERE login = $1 LIMIT 1`

	row := s.pool.QueryRow(ctx, query, userLogin)

	var u models.User
	err := row.Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%w with ID: %s", ErrUserNotFound, userLogin)
		}

		return models.User{}, fmt.Errorf("failed to scan a response row: %w", err)
	}

	return u, nil
}

func (s *DBStorage) AddUser(ctx context.Context, userLogin string, userPassword string) (models.User, error) {
	const addUserQuery = `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id, login, password`
	const addBalanceQuery = `INSERT INTO balance (user_id) VALUES ($1)`

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer rollbackTx(ctx, tx, s.logger)

	row := tx.QueryRow(ctx, addUserQuery, userLogin, userPassword)
	var u models.User
	if err := row.Scan(&u.ID, &u.Login, &u.Password); err != nil {
		return models.User{}, fmt.Errorf("failed to scan a response row: %w", err)
	}

	if _, err := tx.Exec(ctx, addBalanceQuery, u.ID); err != nil {
		return models.User{}, fmt.Errorf("failed to insert data: %w", err)
	}

	cErr := tx.Commit(ctx)
	if cErr != nil {
		return models.User{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return u, nil
}

func (s *DBStorage) GetOrdersByUserID(ctx context.Context) ([]models.Order, error) {
	const query = `
		SELECT number, status, accrual, uploaded_at, user_id
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at ASC
	`

	orders := []models.Order{}

	rows, err := s.pool.Query(ctx, query, ctx.Value(common.KeyUserID))
	if err != nil {
		return []models.Order{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt, &o.UserID)
		if err != nil {
			return []models.Order{}, fmt.Errorf("failed to scan query: %w", err)
		}

		orders = append(orders, o)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return []models.Order{}, fmt.Errorf("failed to read query: %w", err)
	}

	return orders, nil
}

func (s *DBStorage) GetOrdersByStatus(ctx context.Context, statuses ...string) ([]models.Order, error) {
	const query = `SELECT number, status, accrual, uploaded_at, user_id FROM orders WHERE status = any($1)`

	orders := []models.Order{}

	rows, err := s.pool.Query(ctx, query, statuses)
	if err != nil {
		return []models.Order{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var o models.Order
		err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt, &o.UserID)
		if err != nil {
			return []models.Order{}, fmt.Errorf("failed to scan query: %w", err)
		}

		orders = append(orders, o)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return []models.Order{}, fmt.Errorf("failed to read query: %w", err)
	}

	return orders, nil
}

func (s *DBStorage) AddOrder(ctx context.Context, number string) (models.Order, bool, error) {
	const query = `
		WITH new_order AS (
			INSERT INTO orders (number, user_id) VALUES ($1, $2)
			ON CONFLICT (number) DO NOTHING
			RETURNING *
		)
		SELECT number, status, accrual, uploaded_at, user_id, true as is_new FROM new_order
		UNION
		SELECT number, status, accrual, uploaded_at, user_id, false as is_new FROM orders WHERE number = $1
	`
	row := s.pool.QueryRow(ctx, query, number, ctx.Value(common.KeyUserID))

	var o models.Order
	var isNewOrder bool

	err := row.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt, &o.UserID, &isNewOrder)
	if err != nil {
		return o, false, fmt.Errorf("failed to scan a response row: %w", err)
	}

	return o, isNewOrder, nil
}

func (s *DBStorage) UpdateOrder(ctx context.Context, number string, status string, accrual int) error {
	const updateQuery = `UPDATE orders SET (status, accrual) = ($2, $3) WHERE number = $1 RETURNING user_id`
	const updateBalanceQuery = `UPDATE balance SET current = current + $1 WHERE user_id = $2`

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer rollbackTx(ctx, tx, s.logger)

	row := tx.QueryRow(ctx, updateQuery, number, status, accrual)
	var userID int
	if err := row.Scan(&userID); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if status == "PROCESSED" {
		if _, err := tx.Exec(ctx, updateBalanceQuery, accrual, userID); err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}
	}

	cErr := tx.Commit(ctx)
	if cErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *DBStorage) GetWithdrawals(ctx context.Context) ([]models.Withdraw, error) {
	const query = `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at ASC
	`

	withdrawals := []models.Withdraw{}

	rows, err := s.pool.Query(ctx, query, ctx.Value(common.KeyUserID))
	if err != nil {
		return []models.Withdraw{}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w models.Withdraw
		err = rows.Scan(&w.OrderNumber, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return []models.Withdraw{}, fmt.Errorf("failed to scan query: %w", err)
		}

		withdrawals = append(withdrawals, w)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return []models.Withdraw{}, fmt.Errorf("failed to read query: %w", err)
	}

	return withdrawals, nil
}

func (s *DBStorage) AddWithdraw(ctx context.Context, orderNumber string, sum int) error {
	const getBalanceQuery = `SELECT current FROM balance WHERE user_id = $1 LIMIT 1`
	const addQuery = `
		INSERT INTO withdrawals (order_number, sum, user_id) VALUES ($1, $2, $3) 
		RETURNING order_number, sum, processed_at
	`
	const updateBalanceQuery = `UPDATE balance SET (current, withdrawn) = (current - $1, withdrawn + $1) WHERE user_id = $2`

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer rollbackTx(ctx, tx, s.logger)

	row := tx.QueryRow(ctx, getBalanceQuery, ctx.Value(common.KeyUserID))

	var current int
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("failed to scan a response row: %w", err)
	}

	if current < sum {
		return ErrUserInsufficientFunds
	}

	if _, err := tx.Exec(ctx, addQuery, orderNumber, sum, ctx.Value(common.KeyUserID)); err != nil {
		return fmt.Errorf("failed to add withdrawn: %w", err)
	}

	if _, err := tx.Exec(ctx, updateBalanceQuery, sum, ctx.Value(common.KeyUserID)); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	cErr := tx.Commit(ctx)
	if cErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *DBStorage) GetBalance(ctx context.Context) (models.Balance, error) {
	const query = `SELECT current, withdrawn FROM balance WHERE user_id = $1 LIMIT 1`
	row := s.pool.QueryRow(ctx, query, ctx.Value(common.KeyUserID))

	var b models.Balance
	err := row.Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		return models.Balance{}, fmt.Errorf("failed to scan a response row: %w", err)
	}

	return b, nil
}

func (s *DBStorage) Ping(ctx context.Context) error {
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	return nil
}

func (s *DBStorage) Close() error {
	s.pool.Close()
	return nil
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func initPool(ctx context.Context, logger *zap.Logger, dbURI string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	poolCfg.ConnConfig.Tracer = &queryTracer{logger: logger}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}

	return pool, nil
}

func rollbackTx(ctx context.Context, tx pgx.Tx, logger *zap.Logger) {
	if rErr := tx.Rollback(ctx); rErr != nil {
		if !errors.Is(rErr, pgx.ErrTxClosed) {
			logger.Error("failed to rollback the transaction", zap.Error(rErr))
		}
	}
}

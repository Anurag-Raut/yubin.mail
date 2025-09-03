package store

import (
	"context"
	"fmt"

	"github.com/Yubin-email/smtp-server/logger"
	"github.com/jackc/pgx/v5"
)

var db *pgx.Conn

func InitStore() error {
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:5432/%s", "smtpuser", "smtppass", "smtp")
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	db = conn

	return nil
}

func StoreEmail(from string, to string, data string) error {

	_, err := db.Exec(
		context.Background(),
		`INSERT INTO emails ("from", "to", data, mailbox)
     VALUES ($1, $2, $3, $4)`,
		from, to, data, "inbox",
	)
	logger.Println("INSERTED INTO THE STORE")
	if err != nil {
		return fmt.Errorf("insert error: %w", err)
	}
	return nil
}

func CloseStore() error {
	if db != nil {
		return db.Close(context.Background())
	}
	return nil
}

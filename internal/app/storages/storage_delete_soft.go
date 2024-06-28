package storages

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var timeSleep = 24 * time.Hour

func deleteSoftDeletedRecords(ctx context.Context, db *pgxpool.Pool) {
	for {
		query := `DELETE FROM short_urls WHERE is_deleted=true`
		_, err := db.Exec(ctx, query)
		if err != nil {
			log.Printf("Ошибка при удалении мягко удалённых записей: %v", err)
		}

		time.Sleep(timeSleep)
	}
}

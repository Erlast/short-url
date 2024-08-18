package components

import (
	"context"
	"log"
	"time"

	"github.com/Erlast/short-url.git/internal/app/storages"
)

var timeSleep = 24 * time.Hour

func DeleteSoftDeletedRecords(ctx context.Context, store storages.URLStorage) {
	for {
		err := store.DeleteHard(ctx)
		if err != nil {
			log.Printf("Ошибка работы команды %v", err)
		}

		time.Sleep(timeSleep)
	}
}

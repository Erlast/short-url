package components

import (
	"context"
	"log"
	"time"

	"github.com/Erlast/short-url.git/internal/app/storages"
)

// timeSleep интервал запуска компонента
var timeSleep = 24 * time.Hour

// DeleteSoftDeletedRecords функция удаления записей из харанилища которые ранее были мягко удалены
func DeleteSoftDeletedRecords(ctx context.Context, store storages.URLStorage) {
	for {
		// Удаляем из хранилища
		err := store.DeleteHard(ctx)
		if err != nil {
			log.Printf("Ошибка работы команды %v", err)
		}

		time.Sleep(timeSleep)
	}
}

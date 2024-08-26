package components

// Func TestDeleteSoftDeletedRecords(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	store := storages.NewMockURLStorage(ctrl)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	store.EXPECT().DeleteHard(ctx).Return(nil).Times(1)
//
//	done := make(chan bool)
//
//	go func() {
//		DeleteSoftDeletedRecords(ctx, store)
//		done <- true
//	}()
//
//	time.Sleep(2 * time.Millisecond)
//
//	cancel()
//
//	<-done
//
//	assert.True(t, true, "DeleteSoftDeletedRecords завершилась корректно")
//}
//
// func TestDeleteSoftDeletedRecords_Error(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	store := storages.NewMockURLStorage(ctrl)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	store.EXPECT().DeleteHard(ctx).Return(errors.New("mock error")).Times(1)
//
//	done := make(chan bool)
//
//	go func() {
//		DeleteSoftDeletedRecords(ctx, store)
//		done <- true
//	}()
//
//	time.Sleep(2 * time.Millisecond)
//
//	cancel()
//
//	<-done
//
//	assert.True(t, true, "DeleteSoftDeletedRecords завершилась корректно")
// }.

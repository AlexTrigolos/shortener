package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDeleter(t *testing.T) {
	t.Run("удачное создание", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		doneCh := make(chan struct{})
		delChan := make(chan URLsData, 1024)
		defer close(delChan)
		deleter, _ := NewDeleter(doneCh, urls, zap.NewNop())

		expected := &Deleter{storeURLs: urls, delChan: delChan, logger: zap.NewNop()}
		assert.Equal(t, expected.storeURLs, deleter.storeURLs)
		assert.Equal(t, reflect.TypeOf(expected.delChan), reflect.TypeOf(deleter.delChan))
		assert.Equal(t, expected.logger, deleter.logger)
		close(doneCh)
	})
}

func TestDeleter_AddURLs(t *testing.T) {
	t.Run("удачное добавление на удаление", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)
		doneCh := make(chan struct{})
		deleter, _ := NewDeleter(doneCh, urls, zap.NewNop())

		addURLs := []string{"https://google.com", "https://yandex.ru"}
		userID := "123"
		deleter.AddURLs(addURLs, userID)

		select {
		case urlsData, ex := <-deleter.delChan:
			require.True(t, ex)
			assert.Equal(t, URLsData{Urls: addURLs, UserID: userID}, urlsData)
		case <-time.After(1 * time.Second):
			t.Fatal("не дождались: delChan не имеет ничего за 1 с")
		}
		close(doneCh)
	})
}

func TestDeleter_flushURLS(t *testing.T) {
	t.Run("удачное закрытие обновляет удаленные", func(t *testing.T) {
		t.Parallel()

		urls, err := NewURL()
		require.NoError(t, err)
		err = urls.Set(t.Context(), "short_1", "https://google.com", "user_id")
		require.NoError(t, err)
		err = urls.Set(t.Context(), "short_2", "https://yandex.ru", "user_id")
		require.NoError(t, err)
		err = urls.Set(t.Context(), "short_3", "https://yandex-go.ru", "user_id")
		require.NoError(t, err)
		err = urls.Set(t.Context(), "short_4", "https://yandex-practicum.ru", "user_id_2")
		require.NoError(t, err)
		doneCh := make(chan struct{})
		deleter, finishCh := NewDeleter(doneCh, urls, zap.NewNop())

		deleter.delChan <- URLsData{Urls: []string{"short_1", "short_2"}, UserID: "user_id"}
		time.Sleep(11 * time.Second) // ждем выполнения
		select {
		case <-deleter.delChan:
			t.Fatal("из канала должны были быть забраны данные")
		case <-time.After(1 * time.Second):
			require.Contains(t, urls.deleteFlag, "short_1")
			require.Contains(t, urls.deleteFlag, "short_2")
			assert.Contains(t, urls.data, "short_3")
			assert.Contains(t, urls.data, "short_4")
			assert.Contains(t, urls.values, "https://yandex-go.ru")
			assert.Contains(t, urls.values, "https://yandex-practicum.ru")
			assert.Contains(t, urls.byUserIds, "user_id")
			assert.Contains(t, urls.byUserIds, "user_id_2")
			t.Log("удачно забрали из канала данные и удалили")
		}

		// удстоверяемся, что при закрытии канала закрываются и каналы удалителя
		close(doneCh)
		select {
		case _, ex := <-finishCh:
			assert.False(t, ex)
		case <-time.After(1 * time.Second):
			t.Fatal("не дождались: finishCh не закрыт за 1 с")
		}
		select {
		case _, ex := <-deleter.delChan:
			assert.False(t, ex)
		case <-time.After(1 * time.Second):
			t.Fatal("не дождались: delChan не закрыт за 1 с")
		}
	})
}

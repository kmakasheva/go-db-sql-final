package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.NotEmpty(t, id, "Не удалось выдать идентификатор для добавляемой посылки")

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	RecievedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении только что добавленной посылки")
	require.Equal(t, parcel, RecievedParcel, "Инфо о посылке не совпадает")
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err, "Ошибка при удалении посылки")
	_, err = store.Get(id)
	require.Error(t, err, "Программа должна была выдать ошибку")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	//настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка при подключении к БД")
	defer db.Close()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.NotEmpty(t, id, "Не удалось выдать идентификатор для добавляемой посылки")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка при изменении адреса")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	ReceivedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении добавленной посылки")
	require.Equal(t, newAddress, ReceivedParcel.Address, "О-оу! Адрес не поменялся(")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка при подключении к БД")
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки в БД")
	require.NotEmpty(t, id, "Не выдался идентификатор")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Ошибка при смене статуса")

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	UpdParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении данных о посылке")
	require.Equal(t, ParcelStatusSent, UpdParcel.Status, "Статус посылки не обновился")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	//настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД")
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка при добавлении посылки в БД")
		require.NotEmpty(t, id, "Пустой идентификатор")
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err, "Ошибка при получении списка посылок из БД")
	require.Equal(t, len(storedParcels), len(parcelMap), "Количество посылок не совпадает")
	// check
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists, "Посылка с идентификатором %d отсутствует в parcelMap", parcel.Number)

		// Проверяем, что все поля совпадают
		require.Equal(t, expectedParcel, parcel, "Значения не совпадают")
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}

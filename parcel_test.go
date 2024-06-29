package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	`github.com/stretchr/testify/assert`
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
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err, "Ошибка открытия БД - %v", err)

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении в БД - %v", err)

	require.NotEmpty(t, id, "Ошибка в получении id - %v", id)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	parcels, err := store.Get(id)
	parcel.Number = parcels.Number
	require.NoError(t, err, "Ошибка получения по id - %v", err)

	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	assert.Equal(t, parcels, parcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	err = store.Delete(id)
	require.NoError(t, err, "Ошибка удаления id - %v", err)

	// проверьте, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	require.Error(t, err, "Ошибка по id - %v не удалена %v", id, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err, "Ошибка открытия БД - %v", err)

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении в БД - %v", err)

	require.NotEmpty(t, id, "Ошибка в получении id - %v", id)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка при изменении адреса - %v", err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	parcels, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения по id - %v", err)

	assert.Equal(t, parcels.Address, newAddress, "Ошибка parcels.Address - %v newAddress - %v", parcels.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err, "Ошибка открытия БД - %v", err)

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении в БД - %v", err)

	require.NotEmpty(t, id, "Ошибка в получении id - %v", id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Ошибка при изменении статуса - %v", err)

	err = store.SetStatus(id, ParcelStatusDelivered)
	require.NoError(t, err, "Ошибка при изменении статуса - %v", err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	parcels, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения по id - %v", err)

	assert.Equal(t, parcels.Status, ParcelStatusDelivered, "Ошибка parcels.Status - %v ParcelStatusDelivered - %v", parcels.Status, ParcelStatusDelivered)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "./tracker.db")
	require.NoError(t, err, "Ошибка открытия БД - %v", err)

	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

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
		require.NoError(t, err, "Ошибка при добавлении в БД - %v", err)

		require.NotEmpty(t, id, "Ошибка в получении id - %v", id) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка получения по id - %v", err) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(parcelMap), len(storedParcels))
	require.NotEmpty(t, storedParcels)

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		require.NotEmpty(t, parcelMap)
		require.Equal(t, len(parcelMap), len(storedParcels))
		for i := 0; i < len(parcelMap); i++ {
			require.Equal(t, parcel, parcelMap[parcel.Number])
		}

	}
}

package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func setupDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	if err != nil {
		return nil, err
	}
	return db, nil
}

func teardownDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := setupDB()
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.Greater(t, parcel.Number, 0)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	storedParcel, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, parcel, storedParcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	parcel, err = store.Get(parcel.Number)
	require.Error(t, err)
	assert.NotNil(t, parcel)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := setupDB()
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotNil(t, parcelID)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcelID, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	updatedParcel, err := store.Get(parcelID)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := setupDB()
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcelID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotNil(t, parcelID)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(parcelID, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	updatedParcel, err := store.Get(parcelID)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, updatedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := setupDB()
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки

	}

	/// get by client
	storedParcels, err := store.GetByClient(client)
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	require.NoError(t, err)
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(parcels), len(storedParcels), "кол-во не совпадает")

	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// убедитесь, что все посылки из storedParcels есть в parcelMap
	// убедитесь, что значения полей полученных посылок заполнены верно
	assert.ElementsMatch(t, parcels, storedParcels)
}

package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47GetBalanceJson = `
{
	"method": "get_balance"
}
`

func TestHandleGetBalanceEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, int64(21000), publishedResponse.Result.(*getBalanceResponse).Balance)
	assert.Nil(t, publishedResponse.Error)
}

func TestHandleGetBalanceEvent_IsolatedApp_NoTransactions(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, int64(0), publishedResponse.Result.(*getBalanceResponse).Balance)
	assert.Nil(t, publishedResponse.Error)
}
func TestHandleGetBalanceEvent_IsolatedApp_Transactions(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 1000,
	})
	// create an unrelated transaction, should not count
	svc.DB.Create(&db.Transaction{
		AppId:      nil,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 1000,
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, int64(1000), publishedResponse.Result.(*getBalanceResponse).Balance)
	assert.Nil(t, publishedResponse.Error)
}

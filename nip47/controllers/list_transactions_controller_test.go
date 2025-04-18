package controllers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

func TestHandleListTransactionsEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 0,
		"until": 0,
		"limit": 10,
		"offset": 0,
		"type": "incoming"
	}
}
`

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	for i, _ := range tests.MockLNClientTransactions {
		settledAt := time.Unix(*tests.MockLNClientTransactions[i].SettledAt, 0)
		err = svc.DB.Create(&db.Transaction{
			Type:            tests.MockLNClientTransactions[i].Type,
			PaymentRequest:  tests.MockLNClientTransactions[i].Invoice,
			Description:     tests.MockLNClientTransactions[i].Description,
			DescriptionHash: tests.MockLNClientTransactions[i].DescriptionHash,
			Preimage:        &tests.MockLNClientTransactions[i].Preimage,
			PaymentHash:     tests.MockLNClientTransactions[i].PaymentHash,
			AmountMsat:      uint64(tests.MockLNClientTransactions[i].Amount),
			FeeMsat:         uint64(tests.MockLNClientTransactions[i].FeesPaid),
			SettledAt:       &settledAt,
			State:           constants.TRANSACTION_STATE_SETTLED,
			AppId:           &app.ID,
			UpdatedAt:       time.Now().Add(time.Duration(-i) * time.Hour),
		}).Error
		assert.NoError(t, err)
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 2, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
	transaction := publishedResponse.Result.(*listTransactionsResponse).Transactions[0]
	assert.Equal(t, tests.MockLNClientTransactions[0].Type, transaction.Type)
	assert.Equal(t, tests.MockLNClientTransactions[0].Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockLNClientTransactions[0].Description, transaction.Description)
	assert.Equal(t, tests.MockLNClientTransactions[0].DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockLNClientTransactions[0].Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockLNClientTransactions[0].PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockLNClientTransactions[0].Amount, transaction.Amount)
	assert.Equal(t, tests.MockLNClientTransactions[0].FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockLNClientTransactions[0].SettledAt, transaction.SettledAt)
	assert.Equal(t, "settled", transaction.State)
}

func TestHandleListTransactionsEvent_UnpaidOutgoingOnly(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 0,
		"until": 0,
		"limit": 10,
		"offset": 0,
		"unpaid_outgoing": true
	}
}
`

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_OUTGOING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 1, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
	transaction := publishedResponse.Result.(*listTransactionsResponse).Transactions[0]
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
}

func TestHandleListTransactionsEvent_UnpaidIncomingOnly(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 0,
		"until": 0,
		"limit": 10,
		"offset": 0,
		"unpaid_incoming": true
	}
}
`

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_OUTGOING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 1, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
	transaction := publishedResponse.Result.(*listTransactionsResponse).Transactions[0]
	assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, transaction.Type)
}

func TestHandleListTransactionsEvent_Unpaid(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 0,
		"until": 0,
		"limit": 10,
		"offset": 0,
		"unpaid": true
	}
}
`

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_OUTGOING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 2, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
}

func TestHandleListTransactionsEvent_Paid(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 0,
		"until": 0,
		"limit": 10,
		"offset": 0
	}
}
`

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	err = svc.DB.Create(&db.Transaction{
		Type:  constants.TRANSACTION_TYPE_OUTGOING,
		State: constants.TRANSACTION_STATE_PENDING,
	}).Error
	assert.NoError(t, err)

	settledPaymentHash := "dummy payment hash"

	err = svc.DB.Create(&db.Transaction{
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		State:       constants.TRANSACTION_STATE_SETTLED,
		PaymentHash: settledPaymentHash,
	}).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 1, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
	transaction := publishedResponse.Result.(*listTransactionsResponse).Transactions[0]
	assert.Equal(t, settledPaymentHash, transaction.PaymentHash)
}

// TODO: add tests for pagination args

package postgres

import (
	"context"
	"fmt"

	"github.com/ShuzoShinagawa1102/offenro/internal/core/fulfillment"
	"github.com/ShuzoShinagawa1102/offenro/internal/core/model"
	postgresdb "github.com/ShuzoShinagawa1102/offenro/internal/platform/postgres/generated"
)

var _ fulfillment.Repository = (*Store)(nil)

func (s *Store) FindPurchase(ctx context.Context, purchaseID string) (model.Purchase, error) {
	return s.loadPurchase(ctx, purchaseID, fulfillment.ErrNotFound)
}

func (s *Store) PrepareFulfillments(ctx context.Context, values []model.MerchantFulfillment) error {
	if len(values) == 0 {
		return nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin prepare fulfillment transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := s.queries.WithTx(tx)
	for _, value := range values {
		_, err := queries.PrepareMerchantFulfillment(ctx, postgresdb.PrepareMerchantFulfillmentParams{
			FulfillmentID: value.ID, PurchaseItemID: value.PurchaseItemID,
			Status: string(value.Status), IdempotencyKey: value.IdempotencyKey,
			DetailsSnapshot: value.DetailsSnapshot,
			CreatedAt:       timestamp(value.CreatedAt), UpdatedAt: timestamp(value.UpdatedAt),
		})
		if err != nil {
			return fmt.Errorf("prepare merchant fulfillment: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit prepared fulfillments: %w", err)
	}
	return nil
}

func (s *Store) ListFulfillments(ctx context.Context, purchaseID string) ([]model.MerchantFulfillment, error) {
	return s.listFulfillments(ctx, purchaseID)
}

func (s *Store) listFulfillments(ctx context.Context, purchaseID string) ([]model.MerchantFulfillment, error) {
	rows, err := s.queries.ListMerchantFulfillments(ctx, purchaseID)
	if err != nil {
		return nil, fmt.Errorf("list merchant fulfillments: %w", err)
	}
	values := make([]model.MerchantFulfillment, 0, len(rows))
	for _, row := range rows {
		values = append(values, model.MerchantFulfillment{
			ID: row.FulfillmentID, PurchaseItemID: row.PurchaseItemID,
			Status: model.MerchantFulfillmentStatus(row.Status), IdempotencyKey: row.IdempotencyKey,
			MerchantOrderRef: textValue(row.MerchantOrderRef), DetailsSnapshot: row.DetailsSnapshot,
			ResponseSnapshot: row.ResponseSnapshot, FailureCode: textValue(row.FailureCode),
			CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		})
	}
	return values, nil
}

func (s *Store) UpdateFulfillment(
	ctx context.Context,
	value model.MerchantFulfillment,
	expectedStatus model.MerchantFulfillmentStatus,
) error {
	_, err := s.queries.UpdateMerchantFulfillment(ctx, postgresdb.UpdateMerchantFulfillmentParams{
		Status: string(value.Status), MerchantOrderRef: optionalText(value.MerchantOrderRef),
		ResponseSnapshot: value.ResponseSnapshot, FailureCode: optionalText(value.FailureCode),
		UpdatedAt: timestamp(value.UpdatedAt), FulfillmentID: value.ID,
		ExpectedStatus: string(expectedStatus),
	})
	if err != nil {
		return fmt.Errorf("update merchant fulfillment: %w", err)
	}
	return nil
}

func (s *Store) UpdatePurchaseStatus(
	ctx context.Context,
	purchaseID string,
	expectedStatus model.PurchaseStatus,
	status model.PurchaseStatus,
) error {
	_, err := s.queries.UpdatePurchaseStatus(ctx, postgresdb.UpdatePurchaseStatusParams{
		Status: string(status), PurchaseID: purchaseID, ExpectedStatus: string(expectedStatus),
	})
	if err != nil {
		return fmt.Errorf("update purchase status: %w", err)
	}
	return nil
}

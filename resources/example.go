package main

import (
	"context"

	"github.com/pkg/errors"
)

type OrderFinishedUseCase struct{}
type Order struct{}

func (o *OrderFinishedUseCase) runFailureFlow(ctx context.Context, err error, order Order) error {
	return nil
}

func (o *OrderFinishedUseCase) Do(ctx context.Context, order Order) error {
	err := o.beginTransaction(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: beginTransaction")
	}
	err = o.changeStateToFinished(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: change state to finished")
	}
	err = o.recalculateOrderPrice(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: recalculate order price")
	}
	err = o.redeemUserCoupons(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: redeem user's coupons")
	}
	err = o.chargeUserCreditCard(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: charging a user")
	}
	err = o.commitTransaction(ctx, order)
	if err != nil {
		err = o.runFailureFlow(ctx, err, order)
		return errors.Wrap(err, "something goes wrong, step: commit transaction")
	}

	return nil
}

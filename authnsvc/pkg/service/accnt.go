package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/error"
	svcevent "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/util"
)

func (svc *basicAuthNService) CreateAccount(
	ctx context.Context, accnt dto.CreateAccountRequest) (resp dto.CreateAccountResponse) {

	hashedPswd, err := util.HashPassword(accnt.Password)
	if err != nil {
		resp.Err = ce.ErrApplication
		svc.cl.Error(ctx, err)
		return
	}

	roleObj, err := svc.accntrepo.GetRoleByName(ctx, accnt.Role)
	if err != nil {
		resp.Err = err
		return
	}

	uid, err := svc.accntrepo.CreateUser(
		ctx, accnt.Name, accnt.Email, hashedPswd, roleObj)
	resp.UserID = uid
	resp.Err = err

	// if account creation was successful fire account created and create policy events
	if err == nil && uid > 0 {
		eventPublisher := svcevent.NewEventPublisher()
		eventErr := eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventAccountCreated,
			svcevent.EventAccountCreatedPayload{
				AccntID: resp.UserID,
				Role:    accnt.Role}))
		svc.cl.LogIfError(ctx, eventErr)

		eventErr = eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventUpsertPolicy,
			svcevent.EventUpsertPolicyPayload{
				Sub:          fmt.Sprint(uid),
				ResourceType: "accounts",
				ResourceID:   fmt.Sprint(uid),
				Action:       "*"}))
		svc.cl.LogIfError(ctx, eventErr)

		eventErr = eventPublisher.Publish(svc.nc)
		svc.cl.LogIfError(ctx, eventErr)
		if eventErr == nil {
			svc.cl.Debug(ctx, fmt.Sprintf(
				"published events: %v", eventPublisher.GetEventNames()))
		}
	}

	return
}

func (svc *basicAuthNService) DeleteAccount(ctx context.Context, aid uint) (err error) {
	err = svc.accntrepo.DeleteUser(ctx, aid)
	if err != nil {
		svc.cl.Error(ctx, fmt.Sprintf("error while deleting account: %d", aid))
		return
	}
	accntDeletedEvent, eventErr := svcevent.NewEvent(
		ctx, svcevent.EventAccountDeleted, svcevent.EventAccountDeletedPayload{AccntID: aid})
	if eventErr != nil {
		svc.cl.Error(ctx, fmt.Sprintf("error creating event [%v]", eventErr))
		return
	}
	eventErr = accntDeletedEvent.Publish(svc.nc)
	svc.cl.LogIfError(ctx, eventErr)
	if eventErr == nil {
		svc.cl.Debug(ctx, fmt.Sprintf("published events: %s", svcevent.EventAccountDeleted))
	}

	return
}

func (svc *basicAuthNService) ListAccount(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) dto.ListAccountResponse {
	var (
		accnts   []dto.GetAccountResponse
		err      error
		pageInfo *dto.Page
	)

	if len(aids) > 0 {
		accnts, err = svc.accntrepo.ListAccountsByIDs(ctx, aids, qp)
	} else {
		accnts, pageInfo, err = svc.accntrepo.ListUser(ctx, qp)
	}

	if err != nil {
		svc.cl.Error(ctx, fmt.Sprintf("err getting accounts [%v]", err))
		return dto.ListAccountResponse{Err: err}
	}

	return dto.ListAccountResponse{Accounts: accnts, Err: err, PageInfo: pageInfo}
}

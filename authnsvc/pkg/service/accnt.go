package service

import (
	"context"
	"fmt"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/error"
	svcevent "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/event"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/model"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/util"
)

func (svc *basicAuthNService) CreateAccount(
	ctx context.Context, accnt dto.CreateAccountRequest) (resp dto.CreateAccountResponse) {

	hashedPswd, err := util.HashPassword(accnt.Password)
	if err != nil {
		resp.Err = ce.ErrApplication
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
		eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventAccountCreated,
			svcevent.EventAccountCreatedPayload{
				AccntID: resp.UserID,
				Role:    accnt.Role}))

		eventPublisher.AddEvent(svcevent.NewEvent(
			ctx, svcevent.EventUpsertPolicy,
			svcevent.EventUpsertPolicyPayload{
				Sub:          fmt.Sprint(uid),
				ResourceType: "accounts",
				ResourceID:   fmt.Sprint(uid),
				Action:       "*"}))

		eventPublisher.Publish(svc.nc)
	}

	return
}

func (svc *basicAuthNService) DeleteAccount(ctx context.Context, aid uint) (err error) {
	err = svc.accntrepo.DeleteUser(ctx, aid)
	if err != nil {
		fmt.Println("failed deleting account:", aid)
		return
	}
	accntDeletedEvent, eventErr := svcevent.NewEvent(
		ctx, svcevent.EventAccountDeleted, svcevent.EventAccountDeletedPayload{AccntID: aid})
	if eventErr != nil {
		fmt.Println(eventErr)
		return
	}
	accntDeletedEvent.Publish(svc.nc)
	fmt.Println("event published")
	return
}

func (svc *basicAuthNService) ListAccount(ctx context.Context, aids []uint, qp *dto.BasicQueryParam) dto.ListAccountResponse {
	var accntObjs []model.User
	var err error
	if len(aids) > 0 {
		accntObjs, err = svc.accntrepo.ListAccountsByIDs(ctx, aids, qp)
	} else {
		accntObjs, err = svc.accntrepo.ListUser(ctx, qp)
	}

	if err != nil {
		fmt.Println("error getting accounts:", err)
		return dto.ListAccountResponse{Err: err}
	}

	var accnts []dto.GetAccountResponse
	for _, ao := range accntObjs {
		accnts = append(accnts, dto.GetAccountResponse{
			AccountID: ao.ID,
			Name:      ao.Name,
			Email:     ao.Email,
			Role:      ao.Role.Name,
		})
	}
	return dto.ListAccountResponse{Accounts: accnts}
}

package alert

import (
	"context"
	alertv1 "crypto-notifier/gen/alert/v1"
	"crypto-notifier/gen/alert/v1/alertv1connect"
	"log/slog"
	"strconv"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

type AlertServer struct {
	alertv1connect.UnimplementedAlertServiceHandler
	q         *Queries
	notifyCh  chan TriggeredInfo
	RefreshCh chan struct{}
}

// CreateAlert
func (a *AlertServer) CreateAlert(
	ctx context.Context,
	req *connect.Request[alertv1.CreateAlertRequest],
) (*connect.Response[alertv1.CreateAlertResponse], error) {
	userID, err := uuid.Parse(req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	created, err := a.q.CreateAlert(ctx, CreateAlertParams{
		UserID:      userID,
		Symbol:      req.Msg.Symbol,
		TargetPrice: strconv.FormatFloat(req.Msg.TargetPrice, 'f', -1, 64),
		Direction:   req.Msg.Direction,
	})
	if err != nil {
		return nil, err
	}

	a.RefreshCh <- struct{}{}

	return connect.NewResponse(&alertv1.CreateAlertResponse{
		Id: created.ID.String(),
	}), nil
}

// ListAlerts
func (a *AlertServer) ListAlerts(
	ctx context.Context,
	req *connect.Request[alertv1.ListAlertsRequest],
) (*connect.Response[alertv1.ListAlertsResponse], error) {
	userID, err := uuid.Parse(req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	alerts, err := a.q.ListAlertsByUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoAlerts := make([]*alertv1.Alert, len(alerts))
	for i, a := range alerts {

		price, err := strconv.ParseFloat(a.TargetPrice, 64)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		protoAlerts[i] = &alertv1.Alert{
			Id:          a.ID.String(),
			UserId:      a.UserID.String(),
			Symbol:      a.Symbol,
			TargetPrice: price,
			Direction:   a.Direction,
			Active:      a.Active,
		}
	}

	return connect.NewResponse(&alertv1.ListAlertsResponse{
		Alerts: protoAlerts,
	}), nil
}

// DeleteAlert
func (a *AlertServer) DeleteAlert(
	ctx context.Context,
	req *connect.Request[alertv1.DeleteAlertRequest],
) (*connect.Response[alertv1.DeleteAlertResponse], error) {

	alertID, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = a.q.DeleteAlert(ctx, alertID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	a.RefreshCh <- struct{}{}

	return connect.NewResponse(&alertv1.DeleteAlertResponse{}), nil
}

func NewAlertServer(q *Queries, notifyCh chan TriggeredInfo, refreshCh chan struct{}) *AlertServer {
	return &AlertServer{
		q:         q,
		notifyCh:  notifyCh,
		RefreshCh: refreshCh,
	}
}

func (a *AlertServer) SubscribeNotifications(
	ctx context.Context,
	req *connect.Request[alertv1.SubscribeNotificationsRequest],
	stream *connect.ServerStream[alertv1.NotificationsResponse],
) error {
	ch := a.notifyCh

	for {
		select {
		case info := <-ch:
			err := stream.Send(&alertv1.NotificationsResponse{
				AlertId:   info.AlertID,
				Symbol:    info.Symbol,
				Price:     info.Price,
				Direction: info.Direction,
			})
			if err != nil {
				slog.Error("failed to send notification", "error", err)
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}

}

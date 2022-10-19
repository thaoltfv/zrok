package controller

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/edge/rest_management_api_client/service"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type untunnelHandler struct {
	cfg *Config
}

func newUntunnelHandler(cfg *Config) *untunnelHandler {
	return &untunnelHandler{cfg: cfg}
}

func (self *untunnelHandler) Handle(params tunnel.UntunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
	logrus.Infof("untunneling for '%v' (%v)", principal.Email, principal.Token)

	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return tunnel.NewUntunnelInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	edge, err := edgeClient(self.cfg.Ziti)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	svcName := params.Body.Service
	svcZId, err := self.findServiceZId(svcName, edge)
	if err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	var senv *store.Environment
	if envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx); err == nil {
		for _, env := range envs {
			if env.ZId == params.Body.ZitiIdentityID {
				senv = env
				break
			}
		}
		if senv == nil {
			err := errors.Errorf("environment with id '%v' not found for '%v", params.Body.ZitiIdentityID, principal.Email)
			logrus.Error(err)
			return tunnel.NewUntunnelNotFound()
		}
	} else {
		logrus.Errorf("error finding environments for account '%v': %v", principal.Email, err)
		return tunnel.NewUntunnelInternalServerError()
	}

	var ssvc *store.Service
	if svcs, err := str.FindServicesForEnvironment(senv.Id, tx); err == nil {
		for _, svc := range svcs {
			if svc.ZId == svcZId {
				ssvc = svc
				break
			}
		}
		if ssvc == nil {
			err := errors.Errorf("service with id '%v' not found for '%v'", svcZId, principal.Email)
			logrus.Error(err)
			return tunnel.NewUntunnelNotFound()
		}
	} else {
		logrus.Errorf("error finding services for account '%v': %v", principal.Email, err)
		return tunnel.NewUntunnelInternalServerError()
	}

	if err := deleteServiceEdgeRouterPolicy(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteServicePolicyDial(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteServicePolicyBind(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := deleteConfig(svcName, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewTunnelInternalServerError()
	}
	if err := deleteService(svcZId, edge); err != nil {
		logrus.Error(err)
		return tunnel.NewUntunnelInternalServerError()
	}

	logrus.Infof("deallocated service '%v'", svcName)

	ssvc.Active = false
	if err := str.UpdateService(ssvc, tx); err != nil {
		logrus.Errorf("error deactivating service '%v': %v", svcZId, err)
		return tunnel.NewUntunnelInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
		return tunnel.NewUntunnelInternalServerError()
	}

	return tunnel.NewUntunnelOK()
}

func (_ *untunnelHandler) findServiceZId(svcName string, edge *rest_management_api_client.ZitiEdgeManagement) (string, error) {
	filter := fmt.Sprintf("name=\"%v\"", svcName)
	limit := int64(1)
	offset := int64(0)
	listReq := &service.ListServicesParams{
		Filter:  &filter,
		Limit:   &limit,
		Offset:  &offset,
		Context: context.Background(),
	}
	listReq.SetTimeout(30 * time.Second)
	listResp, err := edge.Service.ListServices(listReq, nil)
	if err != nil {
		return "", err
	}
	if len(listResp.Payload.Data) == 1 {
		return *(listResp.Payload.Data[0].ID), nil
	}
	return "", errors.Errorf("service '%v' not found", svcName)
}

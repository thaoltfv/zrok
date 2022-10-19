package controller

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type disableHandler struct {
	cfg *Config
}

func newDisableHandler(cfg *Config) *disableHandler {
	return &disableHandler{cfg: cfg}
}

func (self *disableHandler) Handle(params identity.DisableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction: %v", err)
		return identity.NewDisableInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()
	envId, err := self.checkZitiIdentity(params.Body.Identity, principal, tx)
	if err != nil {
		logrus.Errorf("identity check failed: %v", err)
		return identity.NewDisableUnauthorized()
	}
	edge, err := edgeClient(self.cfg.Ziti)
	if err != nil {
		logrus.Errorf("error getting edge client: %v", err)
		return identity.NewDisableInternalServerError()
	}
	if err := self.removeServicesForEnvironment(envId, tx, edge); err != nil {
		logrus.Errorf("error removing services for environment: %v", err)
		return identity.NewDisableInternalServerError()
	}
	if err := self.removeEnvironment(envId, tx); err != nil {
		logrus.Errorf("error removing environment: %v", err)
		return identity.NewDisableInternalServerError()
	}
	if err := deleteEdgeRouterPolicy(params.Body.Identity, edge); err != nil {
		logrus.Errorf("error deleting edge router policy: %v", err)
		return identity.NewDisableInternalServerError()
	}
	if err := deleteIdentity(params.Body.Identity, edge); err != nil {
		logrus.Errorf("error deleting identity: %v", err)
		return identity.NewDisableInternalServerError()
	}
	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing: %v", err)
	}
	return identity.NewDisableOK()
}

func (self *disableHandler) checkZitiIdentity(id string, principal *rest_model_zrok.Principal, tx *sqlx.Tx) (int, error) {
	envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
	if err != nil {
		return -1, err
	}
	for _, env := range envs {
		if env.ZId == id {
			return env.Id, nil
		}
	}
	return -1, errors.Errorf("no such environment '%v'", id)
}

func (self *disableHandler) removeServicesForEnvironment(envId int, tx *sqlx.Tx, edge *rest_management_api_client.ZitiEdgeManagement) error {
	svcs, err := str.FindServicesForEnvironment(envId, tx)
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		svcName := svc.Name
		logrus.Infof("garbage collecting service '%v'", svcName)
		if err := deleteServiceEdgeRouterPolicy(svcName, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteServicePolicyDial(svcName, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteServicePolicyBind(svcName, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteConfig(svcName, edge); err != nil {
			logrus.Error(err)
		}
		if err := deleteService(svc.ZId, edge); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("removed service '%v'", svc.Name)
	}
	return nil
}

func (self *disableHandler) removeEnvironment(envId int, tx *sqlx.Tx) error {
	if err := str.DeleteEnvironment(envId, tx); err != nil {
		return errors.Wrapf(err, "error deleting environment '%d'", envId)
	}
	return nil
}

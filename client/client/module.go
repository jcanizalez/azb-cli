package client

import (
	"azb/client/models"
	"fmt"
	"net/http"
)

type ModuleClient struct {
	Client *Client
}

func (c *ModuleClient) List(organizationId string, filter string) ([]*models.Module, error) {
	req, err := c.Client.newRequest(http.MethodGet, basePath+fmt.Sprintf("organization/%v/module", organizationId), nil)
	if err != nil {
		return nil, err
	}
	var resp models.GetBodyModule
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *ModuleClient) Create(organizationId string, module models.Module) (*models.Module, error) {
	reqBody := models.PostBodyModule{
		Data: &module,
	}

	req, err := c.Client.newRequest(http.MethodPost, basePath+fmt.Sprintf("organization/%v/module", organizationId), reqBody)
	if err != nil {
		return nil, err
	}
	var resp models.PostBodyModule
	_, err = c.Client.do(req, &resp)
	return resp.Data, err
}

func (c *ModuleClient) Delete(organizationID string, moduleId string) error {
	req, err := c.Client.newRequest(http.MethodDelete, fmt.Sprintf(basePath+"organization/%v/module/%v", organizationID, moduleId), nil)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}

func (c *ModuleClient) Update(organizationId string, module models.Module) error {
	reqBody := models.PostBodyModule{
		Data: &module,
	}

	req, err := c.Client.newRequest(http.MethodPatch, fmt.Sprintf(basePath+"organization/%v/module/%v", organizationId, module.ID), reqBody)
	if err != nil {
		return err
	}

	_, err = c.Client.do(req, nil)
	return err
}

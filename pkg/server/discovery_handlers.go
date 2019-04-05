package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	"bitbucket.org/openbankingteam/conformance-suite/pkg/authentication"
	"bitbucket.org/openbankingteam/conformance-suite/pkg/discovery"
)

type PostDiscoveryModelResponse struct {
	TokenEndpoints                 map[string]string   `json:"token_endpoints"`
	TokenEndpointAuthMethods       map[string][]string `json:"token_endpoint_auth_methods"`
	DefaultTokenEndpointAuthMethod map[string]string   `json:"default_token_endpoint_auth_method"`
	AuthorizationEndpoints         map[string]string   `json:"authorization_endpoints"`
	Issuers                        map[string]string   `json:"issuers"`
}

type validationFailuresResponse struct {
	Error discovery.ValidationFailures `json:"error"`
}

type discoveryHandlers struct {
	webJourney Journey
	logger     *logrus.Entry
}

func newDiscoveryHandlers(webJourney Journey, logger *logrus.Entry) discoveryHandlers {
	return discoveryHandlers{webJourney, logger.WithField("handler", "discoveryHandlers")}
}

func (d discoveryHandlers) setDiscoveryModelHandler(c echo.Context) error {
	ctxLogger := d.logger.WithFields(logrus.Fields{
		"function": "setDiscoveryModelHandler",
	})

	discoveryModel := &discovery.Model{}
	if err := c.Bind(discoveryModel); err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse(err))
	}

	failures, err := d.webJourney.SetDiscoveryModel(discoveryModel)
	if err != nil {
		return c.JSON(http.StatusBadRequest, NewErrorResponse(err))
	}

	if !failures.Empty() {
		return c.JSON(http.StatusBadRequest, validationFailuresResponse{failures})
	}

	failures = discovery.ValidationFailures{}
	response := PostDiscoveryModelResponse{
		TokenEndpoints:                 map[string]string{},
		TokenEndpointAuthMethods:       map[string][]string{},
		DefaultTokenEndpointAuthMethod: map[string]string{},
		AuthorizationEndpoints:         map[string]string{},
		Issuers:                        map[string]string{},
	}
	for discoveryItemIndex, discoveryItem := range discoveryModel.DiscoveryModel.DiscoveryItems {
		key := fmt.Sprintf("schema_version=%s", discoveryItem.APISpecification.SchemaVersion)

		url := discoveryItem.OpenidConfigurationURI
		ctxLogger.WithFields(logrus.Fields{
			"url": url,
		}).Info("GET /.well-known/openid-configuration")
		config, e := authentication.OpenIdConfig(url)
		if e != nil {
			ctxLogger.WithFields(logrus.Fields{
				"url": url,
				"err": e,
			}).Error("Error on /.well-known/openid-configuration")
			failures = append(failures, newOpenidConfigurationURIFailure(discoveryItemIndex, e))
		} else {
			response.TokenEndpoints[key] = config.TokenEndpoint
			response.AuthorizationEndpoints[key] = config.AuthorizationEndpoint
			response.Issuers[key] = config.Issuer
			response.TokenEndpointAuthMethods[key] = authentication.SuiteSupportedAuthMethodsMostSecureFirst
			response.DefaultTokenEndpointAuthMethod[key] = authentication.DefaultAuthMethod(config.TokenEndpointAuthMethodsSupported, d.logger)
		}
	}

	if !failures.Empty() {
		return c.JSON(http.StatusBadRequest, validationFailuresResponse{failures})
	}
	return c.JSON(http.StatusCreated, response)
}

func newOpenidConfigurationURIFailure(discoveryItemIndex int, err error) discovery.ValidationFailure {
	return discovery.ValidationFailure{
		Key:   fmt.Sprintf("DiscoveryModel.DiscoveryItems[%d].OpenidConfigurationURI", discoveryItemIndex),
		Error: err.Error(),
	}
}

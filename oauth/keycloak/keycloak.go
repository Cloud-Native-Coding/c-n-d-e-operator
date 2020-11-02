package keycloak

import (
	"context"
	"strings"

	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	"cnde-operator.cloud-native-coding.dev/oauth"
	"github.com/Nerzal/gocloak/v7"
	"github.com/go-logr/logr"
)

//OAUTHProvider data for Keycloak
type OAUTHProvider struct {
	oauth.BaseOAUTHProvider
	log                logr.Logger
	oauthURL           string
	oauthAdminName     string
	oauthAdminPassword string
	oauthAdminRealm    string

	resourceName string
	ingressHost  string

	oauthClientID        string
	oauthInitialPassword string
}

//NewKeycloakOAUTHProvider creates a new OAUTH Provider with Keycloak
func NewKeycloakOAUTHProvider(config *oauth.OAUTHProviderConfig) *OAUTHProvider {
	p := &OAUTHProvider{
		log:                  config.Log,
		oauthURL:             config.OauthURL,
		oauthAdminName:       config.OauthAdminName,
		oauthAdminPassword:   config.OauthAdminPassword,
		oauthAdminRealm:      config.OauthAdminRealm,
		oauthClientID:        config.OauthClientID,
		oauthInitialPassword: config.OauthInitialPassword,

		ingressHost:  config.IngressHost,
		resourceName: config.ResourceName,
	}
	return p
}

func (r *OAUTHProvider) NewRealm(cr *cndev1alpha1.DevEnv) (string, error) {
	reqLogger := r.log.WithValues("Keycloak.URI", r.oauthURL)
	TRUE := true
	accessTokenLifespan := 23 * 60 * 60

	client := gocloak.NewClient(r.oauthURL)
	ctx := context.Background()
	//reqLogger.Info("Login", "name", r.OauthAdminName, "pw", "****", "realm", r.OauthAdminRealm)
	token, err := client.LoginAdmin(ctx, r.oauthAdminName, r.oauthAdminPassword, r.oauthAdminRealm)
	if err != nil {
		reqLogger.Error(err, "Failed to connect to Keycloak.")
		return "", err
	}

	_, err = client.GetRealm(ctx, token.AccessToken, r.resourceName)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			rp := gocloak.RealmRepresentation{
				Realm:               &r.resourceName,
				Enabled:             &TRUE,
				ID:                  &r.resourceName,
				AccessTokenLifespan: &accessTokenLifespan,
			}
			_, err = client.CreateRealm(ctx, token.AccessToken, rp)
			if err != nil {
				reqLogger.Error(err, "Failed to create Realm.")
				return "", err
			}
		} else {
			reqLogger.Error(err, "Failed to get Realm")
			return "", err
		}
	}
	return r.resourceName, nil
}

func (r *OAUTHProvider) DeleteRealm(cr *cndev1alpha1.DevEnv) error {
	reqLogger := r.log.WithValues("Keycloak.URI", r.oauthURL)

	client := gocloak.NewClient(r.oauthURL)
	ctx := context.Background()

	token, err := client.LoginAdmin(ctx, r.oauthAdminName, r.oauthAdminPassword, r.oauthAdminRealm)
	if err != nil {
		reqLogger.Error(err, "Failed to connect to Keycloak.")
		return err
	}

	err = client.DeleteRealm(ctx, token.AccessToken, r.resourceName)
	if err != nil {
		reqLogger.Error(err, "Failed to delete Realm.")
		return err
	}
	return nil
}

func (r *OAUTHProvider) CreateUser(cr *cndev1alpha1.DevEnv) (string, error) {
	reqLogger := r.log.WithValues("Keycloak.URI", r.oauthURL)
	TRUE := true

	client := gocloak.NewClient(r.oauthURL)
	ctx := context.Background()

	token, err := client.LoginAdmin(ctx, r.oauthAdminName, r.oauthAdminPassword, r.oauthAdminRealm)
	if err != nil {
		reqLogger.Error(err, "Failed to connect to Keycloak.")
		return "", err
	}

	u := gocloak.User{
		Username:      &cr.Name,
		Enabled:       &TRUE,
		Email:         &cr.Spec.UserEmail,
		EmailVerified: &TRUE,
	}

	user, err := client.CreateUser(ctx, token.AccessToken, r.resourceName, u)
	if err != nil {
		return "", err
	}

	err = client.SetPassword(ctx, token.AccessToken, user, r.resourceName, r.oauthInitialPassword, true)
	return user, err
}

func (r *OAUTHProvider) CreateClient(cr *cndev1alpha1.DevEnv) (string, error) {
	reqLogger := r.log.WithValues("Keycloak.URI", r.oauthURL)
	TRUE := true
	FALSE := false

	protocol := "openid-connect"
	url := "https://" + r.ingressHost

	client := gocloak.NewClient(r.oauthURL)
	ctx := context.Background()
	token, err := client.LoginAdmin(ctx, r.oauthAdminName, r.oauthAdminPassword, r.oauthAdminRealm)
	if err != nil {
		reqLogger.Error(err, "Failed to connect to Keycloak.")
		return "", err
	}

	c := gocloak.Client{
		ClientID:            &r.oauthClientID,
		Name:                &r.oauthClientID,
		RootURL:             &url,
		RedirectURIs:        &[]string{"*"},
		Enabled:             &TRUE,
		Protocol:            &protocol,
		PublicClient:        &FALSE,
		StandardFlowEnabled: &TRUE,
	}

	_, err = client.CreateClient(ctx, token.AccessToken, r.resourceName, c)
	if err != nil && !strings.Contains(err.Error(), "409") {
		reqLogger.Error(err, "Failed to createClient.")
		return "", err
	}

	clients, err := client.GetClients(ctx,
		token.AccessToken,
		r.resourceName,
		gocloak.GetClientsParams{
			ClientID: &r.oauthClientID,
		},
	)

	if len(clients) == 0 {
		return "", nil // reconsile
	}

	repr, err := client.GetClientSecret(ctx, token.AccessToken, r.resourceName, *clients[0].ID)
	if err != nil {
		reqLogger.Error(err, "Failed to get client secret.")
	}
	return *repr.Value, nil
}

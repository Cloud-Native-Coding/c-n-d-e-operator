package oauth

import (
	cndev1alpha1 "cnde-operator.cloud-native-coding.dev/api/v1alpha1"
	"github.com/go-logr/logr"
)

type OAUTHProvider interface {
	NewRealm(cr *cndev1alpha1.DevEnv) (string, error)
	DeleteRealm(cr *cndev1alpha1.DevEnv) error
	CreateUser(cr *cndev1alpha1.DevEnv) (string, error)
	CreateClient(cr *cndev1alpha1.DevEnv) (string, error)
}

type BaseOAUTHProvider struct {
}

type OAUTHProviderConfig struct {
	Log                logr.Logger
	OauthURL           string
	OauthAdminName     string
	OauthAdminPassword string
	OauthAdminRealm    string

	ResourceName string
	IngressHost  string

	OauthClientID        string
	OauthInitialPassword string
}

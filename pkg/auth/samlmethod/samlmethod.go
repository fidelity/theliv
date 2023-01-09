/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package samlmethod

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	a "github.com/microsoft/kiota/authentication/go/azure"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/microsoftgraph/msgraph-sdk-go/users/item/memberof"

	"github.com/fidelity/theliv/internal/rbac"
	"github.com/fidelity/theliv/pkg/config"
	"github.com/wangli1030/saml"
	"github.com/wangli1030/saml/samlsp"
)

var sp *samlsp.Middleware
var clientID string
var clientSecret string

func Init() {

	authConfig := config.GetThelivConfig().Auth
	clientID = authConfig.ClientID
	clientSecret = authConfig.ClientSecret

	keyPair, err1 := tls.X509KeyPair(authConfig.Cert, authConfig.Key)
	if err1 != nil {
		panic(err1) // TODO handle error
	}
	keyPair.Leaf, err1 = x509.ParseCertificate(keyPair.Certificate[0])
	if err1 != nil {
		panic(err1) // TODO handle error
	}
	var idpMetadata *saml.EntityDescriptor

	if len(authConfig.IDPMetadata) > 0 {
		idpMetadata, err1 = samlsp.ParseMetadata(authConfig.IDPMetadata)
		if err1 != nil {
			panic(err1) // TODO handle error
		}
	} else if authConfig.IDPMetadataPath != "" {
		data, _ := ioutil.ReadFile(authConfig.IDPMetadataPath)
		idpMetadata, err1 = samlsp.ParseMetadata(data)
		if err1 != nil {
			panic(err1) // TODO handle error
		}

	} else if authConfig.IDPMetadataURL != "" {
		idpMetadataURL, err1 := url.Parse(authConfig.IDPMetadataURL)
		if err1 != nil {
			panic(err1) // TODO handle error
		}
		idpMetadata, err1 = samlsp.FetchMetadata(context.Background(), http.DefaultClient,
			*idpMetadataURL)
		if err1 != nil {
			panic(err1) // TODO handle error
		}
	} else {
		panic("please Provide Idp Metadata")
	}

	rootURL, err1 := url.Parse(authConfig.RootURL)
	if err1 != nil {
		panic(err1) // TODO handle error
	}

	samlSP, _ := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})

	metadataURL, err1 := url.Parse(authConfig.MetadataURL)
	if err1 != nil {
		panic(err1)
	}
	samlSP.ServiceProvider.MetadataURL = *metadataURL

	acrURL, err1 := url.Parse(authConfig.AcrURL)
	if err1 != nil {
		panic(err1)
	}
	samlSP.ServiceProvider.AcsURL = *acrURL

	sloURL, err1 := url.Parse(authConfig.SloURL)
	if err1 != nil {
		panic(err1)
	}
	samlSP.ServiceProvider.SloURL = *sloURL

	samlSP.ServiceProvider.EntityID = authConfig.EntityID
	// //Change session token name
	// cookieSP, ok := samlSP.Session.(samlsp.CookieSessionProvider)
	// if !ok {
	// 	panic("session cookie cast failed")
	// }
	// cookieSP.Name = "thelivToken"
	// samlSP.Session = cookieSP

	sp = samlSP
}
func GetSP() *samlsp.Middleware {
	return sp
}
func CheckAuthorization(r *http.Request) (*http.Request, error) {
	session, err := sp.Session.GetSession(r)
	if err != nil {
		return r, err
	}
	if session != nil {
		r = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
		return r, nil
	}
	return r, errors.New("not this Auth method")
}

func HandleStartAuthFlow(w http.ResponseWriter, r *http.Request) {
	var binding, bindingLocation string
	binding = saml.HTTPRedirectBinding
	bindingLocation = sp.ServiceProvider.GetSSOBindingLocation(binding)

	authReq, err := sp.ServiceProvider.MakeAuthenticationRequest(bindingLocation, binding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//change redirect url to main page
	mainuris := getRedirectUrl(r.Header.Get("redirect"))
	mainuri, err := url.Parse(mainuris)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	r.URL = mainuri
	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way to long. Instead
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.

	relayState, err := sp.RequestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if binding == saml.HTTPRedirectBinding {
		redirectURL, err := authReq.Redirect(relayState, &sp.ServiceProvider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("X-Location", redirectURL.String())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func getRedirectUrl(url string) (redirect string) {
	redirect = "/theliv/"
	path := strings.Split(url, redirect)
	if len(path) == 2 {
		redirect = redirect + path[1]
	}
	return
}

type Samlinfo struct {
}

func (Samlinfo) GetUser(r *http.Request) (*rbac.User, error) {
	if session := samlsp.SessionFromContext(r.Context()); session != nil {
		// this will panic if we have the wrong type of Session, and that is OK.
		emptyResult := []string{""}
		sessionWithAttributes := session.(samlsp.SessionWithAttributes)
		attributes := sessionWithAttributes.GetAttributes()
		surname, ok := attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"]
		if !ok {
			surname = emptyResult
		}
		givenname, ok := attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"]
		if !ok {
			givenname = emptyResult
		}
		displayname, ok := attributes["http://schemas.microsoft.com/identity/claims/displayname"]
		if !ok {
			displayname = emptyResult
		}
		emailaddress, ok := attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]
		if !ok {
			emailaddress = emptyResult
		}
		corpid, ok := attributes["corpid"]
		if !ok {
			corpid = displayname
		}

		return &rbac.User{
			Surname:      surname[0],
			Givenname:    givenname[0],
			UID:          corpid[0],
			Displayname:  displayname[0],
			Emailaddress: emailaddress[0],
		}, nil
	}
	return nil, errors.New("Session is empty")
}

func (Samlinfo) GetADgroups(r *http.Request) ([]string, error) {
	if session := samlsp.SessionFromContext(r.Context()); session != nil {
		// this will panic if we have the wrong type of Session, and that is OK.
		sessionWithAttributes := session.(samlsp.SessionWithAttributes)
		attributes := sessionWithAttributes.GetAttributes()
		adgroups, ok := attributes["http://schemas.microsoft.com/ws/2008/06/identity/claims/groups"]
		if !ok {
			adgrouplink, ok := attributes["http://schemas.microsoft.com/claims/groups.link"]
			if !ok {
				return nil, errors.New("do not get the AD group")
			}
			ads, err := GetADgroupsByLink(adgrouplink[0])
			if err != nil {
				return nil, err
			}
			return ads, nil
		}

		return adgroups, nil
	}
	return nil, errors.New("get wrong session")
}

func GetADgroupsByLink(link string) ([]string, error) {
	//check whether link is recognized
	//change client id and secret as input
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if url.Host != "graph.windows.net" || strings.Split(url.Path, "/")[2] != "users" || strings.Split(url.Path, "/")[4] != "getMemberObjects" {
		return nil, errors.New("unrecognized adgroup link")
	}
	tenantId := strings.Split(url.Path, "/")[1]
	userId := strings.Split(url.Path, "/")[3]
	//THE GO SDK IS IN PREVIEW. NON-PRODUCTION USE ONLY
	cred, err := azidentity.NewClientSecretCredential(
		tenantId,
		clientID,
		clientSecret,
		nil,
	)
	if err != nil {
		return nil, err
	}

	auth, err := a.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		fmt.Printf("Error authentication provider: %v\n", err)
		return nil, err
	}
	requestAdapter, err := msgraphsdk.NewGraphRequestAdapter(auth)
	if err != nil {
		fmt.Printf("Error creating adapter: %v\n", err)
		return nil, err
	}

	graphClient := msgraphsdk.NewGraphServiceClient(requestAdapter)

	requestParameters := &memberof.MemberOfRequestBuilderGetQueryParameters{
		Select: []string{"displayName"},
	}
	options := &memberof.MemberOfRequestBuilderGetOptions{
		Q: requestParameters,
	}
	result, err := graphClient.UsersById(userId).MemberOf().Get(options)
	if err != nil {
		return nil, err
	}
	adgroups := []string{}
	for _, value := range result.GetValue() {
		data := *value.Entity.GetAdditionalData()["displayName"].(*string)
		adgroups = append(adgroups, data)
	}

	for {
		if result.GetNextLink() == nil {
			break
		}
		result, err = memberof.NewMemberOfRequestBuilder(*result.GetNextLink(), requestAdapter).Get(nil)
		if err != nil {
			fmt.Printf("Error getting Adgroups: %v\n", err)
			return nil, err
		}
		for _, value := range result.GetValue() {
			data := *value.Entity.GetAdditionalData()["displayName"].(*string)
			adgroups = append(adgroups, data)
		}
	}
	return adgroups, nil
}

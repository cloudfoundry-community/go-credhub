package credhub_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"testing"

	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	credhub "github.com/jghiloni/credhub-api"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOAuthClient(t *testing.T) {
	spec.Run(t, "OAuth2 Client", func(t *testing.T, when spec.G, it spec.S) {
		cs := mockCredhubServer()

		serverCert := cs.Certificate()

		getClient := func(cu, ci, cs string, skip bool) (*http.Client, error) {
			endpoint, err := credhub.UAAEndpoint(cu, true)
			if err != nil {
				return nil, err
			}

			//var sslcli *http.Client
			var t *http.Transport
			if skip {
				t = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			} else {
				certs, _ := x509.SystemCertPool()
				certs.AddCert(serverCert)
				t = &http.Transport{
					TLSClientConfig: &tls.Config{RootCAs: certs},
				}
			}
			sslcli := &http.Client{Transport: t}

			ctx := context.WithValue(context.TODO(), oauth2.HTTPClient, sslcli)
			cfg := &clientcredentials.Config{
				ClientID:     ci,
				ClientSecret: cs,
				TokenURL:     endpoint.TokenURL,
				Scopes:       []string{"credhub.read", "credhub.write"},
			}

			return cfg.Client(ctx), nil
		}

		it.Before(func() {
			RegisterTestingT(t)
		})

		goodAuthTests := func(client *http.Client, err error) {
			it("should work if credentials are correct", func() {
				Expect(client).To(Not(BeNil()))
				Expect(err).To(Not(HaveOccurred()))
				r, err2 := client.Get(cs.URL + "/some-url")
				Expect(r).To(Not(BeNil()))
				Expect(r.StatusCode).To(Equal(200))
				Expect(err2).To(BeNil())
			})
		}

		badAuthTests := func(client *http.Client, err error) {
			it("should not work if credentials are incorrect", func() {
				Expect(client).To(Not(BeNil()))
				Expect(err).To(Not(HaveOccurred()))
				r, err2 := client.Get(cs.URL + "/some-url")
				Expect(r).To(BeNil())
				urlerr, ok := err2.(*url.Error)
				Expect(ok).To(BeTrue())
				Expect(urlerr).To(Not(BeNil()))
				Expect(strings.HasSuffix(urlerr.Error(), "401 Unauthorized"))
			})
		}

		when("testing with skipping SSL validation", func() {
			goodAuthTests(getClient(cs.URL, "user", "pass", true))
			badAuthTests(getClient(cs.URL, "baduser", "badpass", true))
		})

		when("testing without skipping SSL validation", func() {
			goodAuthTests(getClient(cs.URL, "user", "pass", false))
			badAuthTests(getClient(cs.URL, "baduser", "badpass", false))
		})
	}, spec.Report(report.Terminal{}))
}
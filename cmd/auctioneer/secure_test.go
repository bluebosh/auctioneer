package main_test

import (
	"os"
	"path"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/cfhttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Secure", func() {

	var (
		certPath                      string
		factory                       auctioneer.ClientFactory
		err                           error
		certFile, keyFile, caCertFile string
		tlsConfig                     *auctioneer.TLSConfig
	)

	JustBeforeEach(func() {
		factory, err = auctioneer.NewClientFactory(cfhttp.NewClient(), tlsConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("insecure mode", func() {
		BeforeEach(func() {
			auctioneerProcess = ginkgomon.Invoke(runner)

			certPath = path.Join(os.Getenv("GOPATH"), "src/code.cloudfoundry.org/auctioneer/cmd/auctioneer/fixtures/certs")
			certFile = path.Join(certPath, "client.crt")
			keyFile = path.Join(certPath, "client.key")
			caCertFile = path.Join(certPath, "server-ca.crt")
		})

		Describe("When the auctioneer receives an HTTP request", func() {
			BeforeEach(func() {
				tlsConfig = nil
			})

			It("accepts the connection", func() {
				client := factory.CreateClient(auctioneerAddress)

				Eventually(func() error {
					return client.RequestTaskAuctions([]*auctioneer.TaskStartRequest{})
				}).ShouldNot(HaveOccurred())
			})
		})

		Describe("When the auctioneer receives an HTTPS request", func() {
			Describe("When requireTLS is not set", func() {
				BeforeEach(func() {
					tlsConfig = &auctioneer.TLSConfig{
						RequireTLS: false,
						CertFile:   certFile,
						KeyFile:    keyFile,
						CaCertFile: caCertFile,
					}
				})
				It("accepts the connection", func() {
					client := factory.CreateClient(auctioneerAddress)
					Eventually(func() error {
						return client.RequestTaskAuctions([]*auctioneer.TaskStartRequest{})
					}).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Context("dual mode", func() {
		BeforeEach(func() {
			auctioneerProcess = ginkgomon.Invoke(runner)

			certPath = path.Join(os.Getenv("GOPATH"), "src/code.cloudfoundry.org/auctioneer/cmd/auctioneer/fixtures/certs")
			certFile = path.Join(certPath, "client.crt")
			keyFile = path.Join(certPath, "client.key")
			caCertFile = path.Join(certPath, "server-ca.crt")
		})

		Describe("When the auctioneer receives an HTTP request", func() {
			BeforeEach(func() {
				tlsConfig = nil
			})

			It("accepts the connection", func() {
				client := factory.CreateClient(auctioneerAddress)

				Eventually(func() error {
					return client.RequestTaskAuctions([]*auctioneer.TaskStartRequest{})
				}).ShouldNot(HaveOccurred())
			})
		})

		Describe("When the auctioneer receives an HTTPS request", func() {
			Describe("When requireTLS is not set", func() {
				BeforeEach(func() {
					tlsConfig = &auctioneer.TLSConfig{
						RequireTLS: false,
						CertFile:   certFile,
						KeyFile:    keyFile,
						CaCertFile: caCertFile,
					}
				})
				It("accepts the connection", func() {
					client := factory.CreateClient(auctioneerAddressSecurable)
					Eventually(func() error {
						return client.RequestTaskAuctions([]*auctioneer.TaskStartRequest{})
					}).Should(HaveOccurred())
				})
			})

			Describe("When requireTLS is set", func() {
				BeforeEach(func() {
					tlsConfig = &auctioneer.TLSConfig{
						RequireTLS: true,
						CertFile:   certFile,
						KeyFile:    keyFile,
						CaCertFile: caCertFile,
					}
				})
				It("accepts the connection", func() {
					client := factory.CreateClient(auctioneerAddressSecurable)
					Eventually(func() error {
						return client.RequestTaskAuctions([]*auctioneer.TaskStartRequest{})
					}).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})

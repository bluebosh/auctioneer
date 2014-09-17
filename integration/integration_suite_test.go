package integration_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/auction/communication/nats/auction_nats_client"
	"github.com/cloudfoundry-incubator/auctioneer/integration/auctioneer_runner"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/natsrunner"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"

	"testing"
	"time"
)

var auctioneerPath string
var simulationRepPath string

var dotNetRep, lucidRep *gexec.Session
var dotNetGuid, lucidGuid = "guid-dot-net", "guid-lucid64"
var dotNetStack, lucidStack = "dot-net", "lucid64"
var dotNetPresence, lucidPresence ifrit.Process

var natsPort, etcdPort int

var runner *auctioneer_runner.AuctioneerRunner
var etcdRunner *etcdstorerunner.ETCDClusterRunner
var natsRunner *natsrunner.NATSRunner
var etcdClient storeadapter.StoreAdapter
var bbs *Bbs.BBS
var repClient *auction_nats_client.AuctionNATSClient
var logger lager.Logger

func TestIntegration(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

type BuiltAssets struct {
	AuctioneerPath    string
	SimulationRepPath string
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	assets := BuiltAssets{}
	assets.AuctioneerPath, err = gexec.Build("github.com/cloudfoundry-incubator/auctioneer", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	assets.SimulationRepPath, err = gexec.Build("github.com/cloudfoundry-incubator/auction/simulation/repnode")
	Ω(err).ShouldNot(HaveOccurred())

	marshalledAssets, err := json.Marshal(assets)
	Ω(err).ShouldNot(HaveOccurred())
	return marshalledAssets
}, func(marshalledAssets []byte) {
	assets := BuiltAssets{}
	err := json.Unmarshal(marshalledAssets, &assets)
	Ω(err).ShouldNot(HaveOccurred())

	auctioneerPath = assets.AuctioneerPath
	simulationRepPath = assets.SimulationRepPath

	etcdPort = 5001 + GinkgoParallelNode()
	natsPort = 4001 + GinkgoParallelNode()

	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
	natsRunner = natsrunner.NewNATSRunner(natsPort)

	etcdClient = etcdRunner.Adapter()

	logger = lagertest.NewTestLogger("test")

	bbs = Bbs.NewBBS(etcdClient, timeprovider.NewTimeProvider(), logger)

	runner = auctioneer_runner.New(
		auctioneerPath,
		[]string{fmt.Sprintf("http://127.0.0.1:%d", etcdPort)},
		[]string{fmt.Sprintf("127.0.0.1:%d", natsPort)},
	)
})

var _ = BeforeEach(func() {
	etcdRunner.Start()
	natsRunner.Start()

	var err error

	dotNetRep, dotNetPresence = startSimulationRep(simulationRepPath, dotNetGuid, dotNetStack, natsPort)
	lucidRep, lucidPresence = startSimulationRep(simulationRepPath, lucidGuid, lucidStack, natsPort)
	repClient, err = auction_nats_client.New(natsRunner.MessageBus, time.Second, logger)
	Ω(err).ShouldNot(HaveOccurred())
})

func startSimulationRep(simulationRepPath, guid string, stack string, natsPort int) (*gexec.Session, ifrit.Process) {
	heartbeater := ifrit.Envoke(bbs.NewExecutorHeartbeat(models.ExecutorPresence{
		ExecutorID: guid,
		Stack:      stack,
	}, time.Second))

	session, err := gexec.Start(exec.Command(
		simulationRepPath,
		"-repGuid", guid,
		"-natsAddrs", fmt.Sprintf("127.0.0.1:%d", natsPort),
	), GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())

	Eventually(session, 5).Should(gbytes.Say("rep node listening"))

	return session, heartbeater
}

var _ = AfterEach(func() {
	runner.KillWithFire()
	etcdRunner.Stop()
	natsRunner.Stop()
	dotNetRep.Kill().Wait()
	lucidRep.Kill().Wait()

	dotNetPresence.Signal(os.Interrupt)
	Eventually(dotNetPresence.Wait()).Should(Receive(BeNil()))

	lucidPresence.Signal(os.Interrupt)
	Eventually(lucidPresence.Wait()).Should(Receive(BeNil()))
})

var _ = SynchronizedAfterSuite(func() {
	if etcdRunner != nil {
		etcdRunner.Stop()
	}
	if natsRunner != nil {
		natsRunner.Stop()
	}
	if dotNetRep != nil {
		dotNetRep.Kill().Wait()
	}
	if lucidRep != nil {
		lucidRep.Kill().Wait()
	}
	if runner != nil {
		runner.KillWithFire()
	}
}, func() {
	gexec.CleanupBuildArtifacts()
})

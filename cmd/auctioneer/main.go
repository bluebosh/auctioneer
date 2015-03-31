package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"

	"github.com/cloudfoundry-incubator/auctioneer/auctionmetricemitterdelegate"
	"github.com/cloudfoundry-incubator/auctioneer/auctionrunnerdelegate"

	"github.com/cloudfoundry-incubator/auctioneer/handlers"
	"github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/cf_http"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/bbs/lock_bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/localip"

	"github.com/cloudfoundry-incubator/auction/auctionrunner"
	"github.com/cloudfoundry-incubator/auction/auctiontypes"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"
	"github.com/pivotal-golang/clock"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var etcdCluster = flag.String(
	"etcdCluster",
	"http://127.0.0.1:4001",
	"comma-separated list of etcd addresses (http://ip:port)",
)

var communicationTimeout = flag.Duration(
	"communicationTimeout",
	10*time.Second,
	"Timeout applied to all HTTP requests.",
)

var heartbeatInterval = flag.Duration(
	"heartbeatInterval",
	lock_bbs.HEARTBEAT_INTERVAL,
	"the interval between heartbeats to the lock",
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:9016",
	"host:port to serve auction and LRP stop requests on",
)

const (
	auctionRunnerTimeout      = 10 * time.Second
	auctionRunnerWorkPoolSize = 1000
	dropsondeDestination      = "localhost:3457"
	dropsondeOrigin           = "auctioneer"
	serverProtocol            = "http"
)

func main() {
	cf_debug_server.AddFlags(flag.CommandLine)
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, reconfigurableSink := cf_lager.New("auctioneer")
	initializeDropsonde(logger)
	bbs := initializeBBS(logger)
	auctionRunner := initializeAuctionRunner(bbs, logger)
	auctionServer := initializeAuctionServer(auctionRunner, logger)
	heartbeater := initializeHeartbeater(bbs, logger)

	cf_http.Initialize(*communicationTimeout)

	members := grouper.Members{
		{"auction-runner", auctionRunner},
		{"auction-server", auctionServer},
		{"heartbeater", heartbeater},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	logger.Info("started")

	err := <-monitor.Wait()
	if err != nil {
		logger.Error("exited-with-failure", err)
		os.Exit(1)
	}

	logger.Info("exited")
}

func initializeAuctionRunner(bbs Bbs.AuctioneerBBS, logger lager.Logger) auctiontypes.AuctionRunner {
	httpClient := cf_http.NewClient()
	httpClient.Transport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	}

	delegate := auctionrunnerdelegate.New(httpClient, bbs, logger)
	metricEmitter := auctionmetricemitterdelegate.New()
	return auctionrunner.New(
		delegate,
		metricEmitter,
		clock.NewClock(),
		workpool.NewWorkPool(auctionRunnerWorkPoolSize),
		logger,
	)
}

func initializeBBS(logger lager.Logger) Bbs.AuctioneerBBS {
	etcdAdapter := etcdstoreadapter.NewETCDStoreAdapter(
		strings.Split(*etcdCluster, ","),
		workpool.NewWorkPool(10),
	)

	err := etcdAdapter.Connect()
	if err != nil {
		logger.Fatal("failed-to-connect-to-etcd", err)
	}

	return Bbs.NewAuctioneerBBS(etcdAdapter, clock.NewClock(), logger)
}

func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("failed to initialize dropsonde: %v", err)
	}
}

func initializeAuctionServer(runner auctiontypes.AuctionRunner, logger lager.Logger) ifrit.Runner {
	return http_server.New(*listenAddr, handlers.New(runner, logger))
}

func initializeHeartbeater(bbs Bbs.AuctioneerBBS, logger lager.Logger) ifrit.Runner {
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Fatal("Couldn't generate uuid", err)
	}

	localIP, err := localip.LocalIP()
	if err != nil {
		logger.Fatal("Couldn't determine local IP", err)
	}

	port := strings.Split(*listenAddr, ":")[1]
	address := fmt.Sprintf("%s://%s:%s", serverProtocol, localIP, port)

	auctioneerPresence := models.NewAuctioneerPresence(uuid.String(), address)
	heartbeater, err := bbs.NewAuctioneerLock(auctioneerPresence, *heartbeatInterval)
	if err != nil {
		logger.Fatal("Couldn't create heartbeater", err)
	}

	return heartbeater
}

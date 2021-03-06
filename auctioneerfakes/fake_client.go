// This file was generated by counterfeiter
package auctioneerfakes

import (
	"sync"

	"code.cloudfoundry.org/auctioneer"
	"code.cloudfoundry.org/lager"
)

type FakeClient struct {
	RequestLRPAuctionsStub        func(logger lager.Logger, lrpStart []*auctioneer.LRPStartRequest) error
	requestLRPAuctionsMutex       sync.RWMutex
	requestLRPAuctionsArgsForCall []struct {
		logger   lager.Logger
		lrpStart []*auctioneer.LRPStartRequest
	}
	requestLRPAuctionsReturns struct {
		result1 error
	}
	RequestTaskAuctionsStub        func(logger lager.Logger, tasks []*auctioneer.TaskStartRequest) error
	requestTaskAuctionsMutex       sync.RWMutex
	requestTaskAuctionsArgsForCall []struct {
		logger lager.Logger
		tasks  []*auctioneer.TaskStartRequest
	}
	requestTaskAuctionsReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeClient) RequestLRPAuctions(logger lager.Logger, lrpStart []*auctioneer.LRPStartRequest) error {
	var lrpStartCopy []*auctioneer.LRPStartRequest
	if lrpStart != nil {
		lrpStartCopy = make([]*auctioneer.LRPStartRequest, len(lrpStart))
		copy(lrpStartCopy, lrpStart)
	}
	fake.requestLRPAuctionsMutex.Lock()
	fake.requestLRPAuctionsArgsForCall = append(fake.requestLRPAuctionsArgsForCall, struct {
		logger   lager.Logger
		lrpStart []*auctioneer.LRPStartRequest
	}{logger, lrpStartCopy})
	fake.recordInvocation("RequestLRPAuctions", []interface{}{logger, lrpStartCopy})
	fake.requestLRPAuctionsMutex.Unlock()
	if fake.RequestLRPAuctionsStub != nil {
		return fake.RequestLRPAuctionsStub(logger, lrpStart)
	} else {
		return fake.requestLRPAuctionsReturns.result1
	}
}

func (fake *FakeClient) RequestLRPAuctionsCallCount() int {
	fake.requestLRPAuctionsMutex.RLock()
	defer fake.requestLRPAuctionsMutex.RUnlock()
	return len(fake.requestLRPAuctionsArgsForCall)
}

func (fake *FakeClient) RequestLRPAuctionsArgsForCall(i int) (lager.Logger, []*auctioneer.LRPStartRequest) {
	fake.requestLRPAuctionsMutex.RLock()
	defer fake.requestLRPAuctionsMutex.RUnlock()
	return fake.requestLRPAuctionsArgsForCall[i].logger, fake.requestLRPAuctionsArgsForCall[i].lrpStart
}

func (fake *FakeClient) RequestLRPAuctionsReturns(result1 error) {
	fake.RequestLRPAuctionsStub = nil
	fake.requestLRPAuctionsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) RequestTaskAuctions(logger lager.Logger, tasks []*auctioneer.TaskStartRequest) error {
	var tasksCopy []*auctioneer.TaskStartRequest
	if tasks != nil {
		tasksCopy = make([]*auctioneer.TaskStartRequest, len(tasks))
		copy(tasksCopy, tasks)
	}
	fake.requestTaskAuctionsMutex.Lock()
	fake.requestTaskAuctionsArgsForCall = append(fake.requestTaskAuctionsArgsForCall, struct {
		logger lager.Logger
		tasks  []*auctioneer.TaskStartRequest
	}{logger, tasksCopy})
	fake.recordInvocation("RequestTaskAuctions", []interface{}{logger, tasksCopy})
	fake.requestTaskAuctionsMutex.Unlock()
	if fake.RequestTaskAuctionsStub != nil {
		return fake.RequestTaskAuctionsStub(logger, tasks)
	} else {
		return fake.requestTaskAuctionsReturns.result1
	}
}

func (fake *FakeClient) RequestTaskAuctionsCallCount() int {
	fake.requestTaskAuctionsMutex.RLock()
	defer fake.requestTaskAuctionsMutex.RUnlock()
	return len(fake.requestTaskAuctionsArgsForCall)
}

func (fake *FakeClient) RequestTaskAuctionsArgsForCall(i int) (lager.Logger, []*auctioneer.TaskStartRequest) {
	fake.requestTaskAuctionsMutex.RLock()
	defer fake.requestTaskAuctionsMutex.RUnlock()
	return fake.requestTaskAuctionsArgsForCall[i].logger, fake.requestTaskAuctionsArgsForCall[i].tasks
}

func (fake *FakeClient) RequestTaskAuctionsReturns(result1 error) {
	fake.RequestTaskAuctionsStub = nil
	fake.requestTaskAuctionsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.requestLRPAuctionsMutex.RLock()
	defer fake.requestLRPAuctionsMutex.RUnlock()
	fake.requestTaskAuctionsMutex.RLock()
	defer fake.requestTaskAuctionsMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ auctioneer.Client = new(FakeClient)

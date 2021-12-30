// Code generated by counterfeiter. DO NOT EDIT.
package apifakes

import (
	"sync"

	"dev.nimak.link/s3-copy-controller/controllers/api"
)

type FakeStoreManager struct {
	GetStub        func(api.ConfigData) api.ObjectStore
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 api.ConfigData
	}
	getReturns struct {
		result1 api.ObjectStore
	}
	getReturnsOnCall map[int]struct {
		result1 api.ObjectStore
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeStoreManager) Get(arg1 api.ConfigData) api.ObjectStore {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 api.ConfigData
	}{arg1})
	stub := fake.GetStub
	fakeReturns := fake.getReturns
	fake.recordInvocation("Get", []interface{}{arg1})
	fake.getMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeStoreManager) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *FakeStoreManager) GetCalls(stub func(api.ConfigData) api.ObjectStore) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *FakeStoreManager) GetArgsForCall(i int) api.ConfigData {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeStoreManager) GetReturns(result1 api.ObjectStore) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 api.ObjectStore
	}{result1}
}

func (fake *FakeStoreManager) GetReturnsOnCall(i int, result1 api.ObjectStore) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 api.ObjectStore
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 api.ObjectStore
	}{result1}
}

func (fake *FakeStoreManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeStoreManager) recordInvocation(key string, args []interface{}) {
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

var _ api.StoreManager = new(FakeStoreManager)

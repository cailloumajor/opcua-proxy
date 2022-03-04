// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package opcua

import (
	"context"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"sync"
)

// Ensure, that ClientProviderMock does implement ClientProvider.
// If this is not the case, regenerate this file with moq.
var _ ClientProvider = &ClientProviderMock{}

// ClientProviderMock is a mock implementation of ClientProvider.
//
// 	func TestSomethingThatUsesClientProvider(t *testing.T) {
//
// 		// make and configure a mocked ClientProvider
// 		mockedClientProvider := &ClientProviderMock{
// 			CloseWithContextFunc: func(ctx context.Context) error {
// 				panic("mock out the CloseWithContext method")
// 			},
// 			NamespaceIndexFunc: func(ctx context.Context, nsURI string) (uint16, error) {
// 				panic("mock out the NamespaceIndex method")
// 			},
// 			StateFunc: func() opcua.ConnState {
// 				panic("mock out the State method")
// 			},
// 			SubscribeWithContextFunc: func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
// 				panic("mock out the SubscribeWithContext method")
// 			},
// 		}
//
// 		// use mockedClientProvider in code that requires ClientProvider
// 		// and then make assertions.
//
// 	}
type ClientProviderMock struct {
	// CloseWithContextFunc mocks the CloseWithContext method.
	CloseWithContextFunc func(ctx context.Context) error

	// NamespaceIndexFunc mocks the NamespaceIndex method.
	NamespaceIndexFunc func(ctx context.Context, nsURI string) (uint16, error)

	// StateFunc mocks the State method.
	StateFunc func() opcua.ConnState

	// SubscribeWithContextFunc mocks the SubscribeWithContext method.
	SubscribeWithContextFunc func(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error)

	// calls tracks calls to the methods.
	calls struct {
		// CloseWithContext holds details about calls to the CloseWithContext method.
		CloseWithContext []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// NamespaceIndex holds details about calls to the NamespaceIndex method.
		NamespaceIndex []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// NsURI is the nsURI argument value.
			NsURI string
		}
		// State holds details about calls to the State method.
		State []struct {
		}
		// SubscribeWithContext holds details about calls to the SubscribeWithContext method.
		SubscribeWithContext []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params *opcua.SubscriptionParameters
			// NotifyCh is the notifyCh argument value.
			NotifyCh chan<- *opcua.PublishNotificationData
		}
	}
	lockCloseWithContext     sync.RWMutex
	lockNamespaceIndex       sync.RWMutex
	lockState                sync.RWMutex
	lockSubscribeWithContext sync.RWMutex
}

// CloseWithContext calls CloseWithContextFunc.
func (mock *ClientProviderMock) CloseWithContext(ctx context.Context) error {
	if mock.CloseWithContextFunc == nil {
		panic("ClientProviderMock.CloseWithContextFunc: method is nil but ClientProvider.CloseWithContext was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockCloseWithContext.Lock()
	mock.calls.CloseWithContext = append(mock.calls.CloseWithContext, callInfo)
	mock.lockCloseWithContext.Unlock()
	return mock.CloseWithContextFunc(ctx)
}

// CloseWithContextCalls gets all the calls that were made to CloseWithContext.
// Check the length with:
//     len(mockedClientProvider.CloseWithContextCalls())
func (mock *ClientProviderMock) CloseWithContextCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockCloseWithContext.RLock()
	calls = mock.calls.CloseWithContext
	mock.lockCloseWithContext.RUnlock()
	return calls
}

// NamespaceIndex calls NamespaceIndexFunc.
func (mock *ClientProviderMock) NamespaceIndex(ctx context.Context, nsURI string) (uint16, error) {
	if mock.NamespaceIndexFunc == nil {
		panic("ClientProviderMock.NamespaceIndexFunc: method is nil but ClientProvider.NamespaceIndex was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		NsURI string
	}{
		Ctx:   ctx,
		NsURI: nsURI,
	}
	mock.lockNamespaceIndex.Lock()
	mock.calls.NamespaceIndex = append(mock.calls.NamespaceIndex, callInfo)
	mock.lockNamespaceIndex.Unlock()
	return mock.NamespaceIndexFunc(ctx, nsURI)
}

// NamespaceIndexCalls gets all the calls that were made to NamespaceIndex.
// Check the length with:
//     len(mockedClientProvider.NamespaceIndexCalls())
func (mock *ClientProviderMock) NamespaceIndexCalls() []struct {
	Ctx   context.Context
	NsURI string
} {
	var calls []struct {
		Ctx   context.Context
		NsURI string
	}
	mock.lockNamespaceIndex.RLock()
	calls = mock.calls.NamespaceIndex
	mock.lockNamespaceIndex.RUnlock()
	return calls
}

// State calls StateFunc.
func (mock *ClientProviderMock) State() opcua.ConnState {
	if mock.StateFunc == nil {
		panic("ClientProviderMock.StateFunc: method is nil but ClientProvider.State was just called")
	}
	callInfo := struct {
	}{}
	mock.lockState.Lock()
	mock.calls.State = append(mock.calls.State, callInfo)
	mock.lockState.Unlock()
	return mock.StateFunc()
}

// StateCalls gets all the calls that were made to State.
// Check the length with:
//     len(mockedClientProvider.StateCalls())
func (mock *ClientProviderMock) StateCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockState.RLock()
	calls = mock.calls.State
	mock.lockState.RUnlock()
	return calls
}

// SubscribeWithContext calls SubscribeWithContextFunc.
func (mock *ClientProviderMock) SubscribeWithContext(ctx context.Context, params *opcua.SubscriptionParameters, notifyCh chan<- *opcua.PublishNotificationData) (Subscription, error) {
	if mock.SubscribeWithContextFunc == nil {
		panic("ClientProviderMock.SubscribeWithContextFunc: method is nil but ClientProvider.SubscribeWithContext was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Params   *opcua.SubscriptionParameters
		NotifyCh chan<- *opcua.PublishNotificationData
	}{
		Ctx:      ctx,
		Params:   params,
		NotifyCh: notifyCh,
	}
	mock.lockSubscribeWithContext.Lock()
	mock.calls.SubscribeWithContext = append(mock.calls.SubscribeWithContext, callInfo)
	mock.lockSubscribeWithContext.Unlock()
	return mock.SubscribeWithContextFunc(ctx, params, notifyCh)
}

// SubscribeWithContextCalls gets all the calls that were made to SubscribeWithContext.
// Check the length with:
//     len(mockedClientProvider.SubscribeWithContextCalls())
func (mock *ClientProviderMock) SubscribeWithContextCalls() []struct {
	Ctx      context.Context
	Params   *opcua.SubscriptionParameters
	NotifyCh chan<- *opcua.PublishNotificationData
} {
	var calls []struct {
		Ctx      context.Context
		Params   *opcua.SubscriptionParameters
		NotifyCh chan<- *opcua.PublishNotificationData
	}
	mock.lockSubscribeWithContext.RLock()
	calls = mock.calls.SubscribeWithContext
	mock.lockSubscribeWithContext.RUnlock()
	return calls
}

// Ensure, that SubscriptionMock does implement Subscription.
// If this is not the case, regenerate this file with moq.
var _ Subscription = &SubscriptionMock{}

// SubscriptionMock is a mock implementation of Subscription.
//
// 	func TestSomethingThatUsesSubscription(t *testing.T) {
//
// 		// make and configure a mocked Subscription
// 		mockedSubscription := &SubscriptionMock{
// 			CancelFunc: func(ctx context.Context) error {
// 				panic("mock out the Cancel method")
// 			},
// 			MonitorWithContextFunc: func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
// 				panic("mock out the MonitorWithContext method")
// 			},
// 		}
//
// 		// use mockedSubscription in code that requires Subscription
// 		// and then make assertions.
//
// 	}
type SubscriptionMock struct {
	// CancelFunc mocks the Cancel method.
	CancelFunc func(ctx context.Context) error

	// MonitorWithContextFunc mocks the MonitorWithContext method.
	MonitorWithContextFunc func(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// Cancel holds details about calls to the Cancel method.
		Cancel []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// MonitorWithContext holds details about calls to the MonitorWithContext method.
		MonitorWithContext []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Ts is the ts argument value.
			Ts ua.TimestampsToReturn
			// Items is the items argument value.
			Items []*ua.MonitoredItemCreateRequest
		}
	}
	lockCancel             sync.RWMutex
	lockMonitorWithContext sync.RWMutex
}

// Cancel calls CancelFunc.
func (mock *SubscriptionMock) Cancel(ctx context.Context) error {
	if mock.CancelFunc == nil {
		panic("SubscriptionMock.CancelFunc: method is nil but Subscription.Cancel was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockCancel.Lock()
	mock.calls.Cancel = append(mock.calls.Cancel, callInfo)
	mock.lockCancel.Unlock()
	return mock.CancelFunc(ctx)
}

// CancelCalls gets all the calls that were made to Cancel.
// Check the length with:
//     len(mockedSubscription.CancelCalls())
func (mock *SubscriptionMock) CancelCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockCancel.RLock()
	calls = mock.calls.Cancel
	mock.lockCancel.RUnlock()
	return calls
}

// MonitorWithContext calls MonitorWithContextFunc.
func (mock *SubscriptionMock) MonitorWithContext(ctx context.Context, ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
	if mock.MonitorWithContextFunc == nil {
		panic("SubscriptionMock.MonitorWithContextFunc: method is nil but Subscription.MonitorWithContext was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Ts    ua.TimestampsToReturn
		Items []*ua.MonitoredItemCreateRequest
	}{
		Ctx:   ctx,
		Ts:    ts,
		Items: items,
	}
	mock.lockMonitorWithContext.Lock()
	mock.calls.MonitorWithContext = append(mock.calls.MonitorWithContext, callInfo)
	mock.lockMonitorWithContext.Unlock()
	return mock.MonitorWithContextFunc(ctx, ts, items...)
}

// MonitorWithContextCalls gets all the calls that were made to MonitorWithContext.
// Check the length with:
//     len(mockedSubscription.MonitorWithContextCalls())
func (mock *SubscriptionMock) MonitorWithContextCalls() []struct {
	Ctx   context.Context
	Ts    ua.TimestampsToReturn
	Items []*ua.MonitoredItemCreateRequest
} {
	var calls []struct {
		Ctx   context.Context
		Ts    ua.TimestampsToReturn
		Items []*ua.MonitoredItemCreateRequest
	}
	mock.lockMonitorWithContext.RLock()
	calls = mock.calls.MonitorWithContext
	mock.lockMonitorWithContext.RUnlock()
	return calls
}

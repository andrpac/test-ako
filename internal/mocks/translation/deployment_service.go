// Code generated by mockery. DO NOT EDIT.

package translation

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1 "github.com/mongodb/mongodb-atlas-kubernetes/v2/api/v1"
	deployment "github.com/mongodb/mongodb-atlas-kubernetes/v2/internal/translation/deployment"
)

// DeploymentServiceMock is an autogenerated mock type for the DeploymentService type
type DeploymentServiceMock struct {
	mock.Mock
}

type DeploymentServiceMock_Expecter struct {
	mock *mock.Mock
}

func (_m *DeploymentServiceMock) EXPECT() *DeploymentServiceMock_Expecter {
	return &DeploymentServiceMock_Expecter{mock: &_m.Mock}
}

// ClusterExists provides a mock function with given fields: ctx, projectID, clusterName
func (_m *DeploymentServiceMock) ClusterExists(ctx context.Context, projectID string, clusterName string) (bool, error) {
	ret := _m.Called(ctx, projectID, clusterName)

	if len(ret) == 0 {
		panic("no return value specified for ClusterExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, projectID, clusterName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, projectID, clusterName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, projectID, clusterName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_ClusterExists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterExists'
type DeploymentServiceMock_ClusterExists_Call struct {
	*mock.Call
}

// ClusterExists is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
//   - clusterName string
func (_e *DeploymentServiceMock_Expecter) ClusterExists(ctx interface{}, projectID interface{}, clusterName interface{}) *DeploymentServiceMock_ClusterExists_Call {
	return &DeploymentServiceMock_ClusterExists_Call{Call: _e.mock.On("ClusterExists", ctx, projectID, clusterName)}
}

func (_c *DeploymentServiceMock_ClusterExists_Call) Run(run func(ctx context.Context, projectID string, clusterName string)) *DeploymentServiceMock_ClusterExists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *DeploymentServiceMock_ClusterExists_Call) Return(_a0 bool, _a1 error) *DeploymentServiceMock_ClusterExists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_ClusterExists_Call) RunAndReturn(run func(context.Context, string, string) (bool, error)) *DeploymentServiceMock_ClusterExists_Call {
	_c.Call.Return(run)
	return _c
}

// ClusterWithProcessArgs provides a mock function with given fields: ctx, cluster
func (_m *DeploymentServiceMock) ClusterWithProcessArgs(ctx context.Context, cluster *deployment.Cluster) error {
	ret := _m.Called(ctx, cluster)

	if len(ret) == 0 {
		panic("no return value specified for ClusterWithProcessArgs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *deployment.Cluster) error); ok {
		r0 = rf(ctx, cluster)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeploymentServiceMock_ClusterWithProcessArgs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClusterWithProcessArgs'
type DeploymentServiceMock_ClusterWithProcessArgs_Call struct {
	*mock.Call
}

// ClusterWithProcessArgs is a helper method to define mock.On call
//   - ctx context.Context
//   - cluster *deployment.Cluster
func (_e *DeploymentServiceMock_Expecter) ClusterWithProcessArgs(ctx interface{}, cluster interface{}) *DeploymentServiceMock_ClusterWithProcessArgs_Call {
	return &DeploymentServiceMock_ClusterWithProcessArgs_Call{Call: _e.mock.On("ClusterWithProcessArgs", ctx, cluster)}
}

func (_c *DeploymentServiceMock_ClusterWithProcessArgs_Call) Run(run func(ctx context.Context, cluster *deployment.Cluster)) *DeploymentServiceMock_ClusterWithProcessArgs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*deployment.Cluster))
	})
	return _c
}

func (_c *DeploymentServiceMock_ClusterWithProcessArgs_Call) Return(_a0 error) *DeploymentServiceMock_ClusterWithProcessArgs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DeploymentServiceMock_ClusterWithProcessArgs_Call) RunAndReturn(run func(context.Context, *deployment.Cluster) error) *DeploymentServiceMock_ClusterWithProcessArgs_Call {
	_c.Call.Return(run)
	return _c
}

// CreateDeployment provides a mock function with given fields: ctx, _a1
func (_m *DeploymentServiceMock) CreateDeployment(ctx context.Context, _a1 deployment.Deployment) (deployment.Deployment, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for CreateDeployment")
	}

	var r0 deployment.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment) (deployment.Deployment, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment) deployment.Deployment); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(deployment.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, deployment.Deployment) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_CreateDeployment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateDeployment'
type DeploymentServiceMock_CreateDeployment_Call struct {
	*mock.Call
}

// CreateDeployment is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 deployment.Deployment
func (_e *DeploymentServiceMock_Expecter) CreateDeployment(ctx interface{}, _a1 interface{}) *DeploymentServiceMock_CreateDeployment_Call {
	return &DeploymentServiceMock_CreateDeployment_Call{Call: _e.mock.On("CreateDeployment", ctx, _a1)}
}

func (_c *DeploymentServiceMock_CreateDeployment_Call) Run(run func(ctx context.Context, _a1 deployment.Deployment)) *DeploymentServiceMock_CreateDeployment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(deployment.Deployment))
	})
	return _c
}

func (_c *DeploymentServiceMock_CreateDeployment_Call) Return(_a0 deployment.Deployment, _a1 error) *DeploymentServiceMock_CreateDeployment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_CreateDeployment_Call) RunAndReturn(run func(context.Context, deployment.Deployment) (deployment.Deployment, error)) *DeploymentServiceMock_CreateDeployment_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteDeployment provides a mock function with given fields: ctx, _a1
func (_m *DeploymentServiceMock) DeleteDeployment(ctx context.Context, _a1 deployment.Deployment) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for DeleteDeployment")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeploymentServiceMock_DeleteDeployment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteDeployment'
type DeploymentServiceMock_DeleteDeployment_Call struct {
	*mock.Call
}

// DeleteDeployment is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 deployment.Deployment
func (_e *DeploymentServiceMock_Expecter) DeleteDeployment(ctx interface{}, _a1 interface{}) *DeploymentServiceMock_DeleteDeployment_Call {
	return &DeploymentServiceMock_DeleteDeployment_Call{Call: _e.mock.On("DeleteDeployment", ctx, _a1)}
}

func (_c *DeploymentServiceMock_DeleteDeployment_Call) Run(run func(ctx context.Context, _a1 deployment.Deployment)) *DeploymentServiceMock_DeleteDeployment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(deployment.Deployment))
	})
	return _c
}

func (_c *DeploymentServiceMock_DeleteDeployment_Call) Return(_a0 error) *DeploymentServiceMock_DeleteDeployment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DeploymentServiceMock_DeleteDeployment_Call) RunAndReturn(run func(context.Context, deployment.Deployment) error) *DeploymentServiceMock_DeleteDeployment_Call {
	_c.Call.Return(run)
	return _c
}

// DeploymentIsReady provides a mock function with given fields: ctx, projectID, deploymentName
func (_m *DeploymentServiceMock) DeploymentIsReady(ctx context.Context, projectID string, deploymentName string) (bool, error) {
	ret := _m.Called(ctx, projectID, deploymentName)

	if len(ret) == 0 {
		panic("no return value specified for DeploymentIsReady")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, projectID, deploymentName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, projectID, deploymentName)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, projectID, deploymentName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_DeploymentIsReady_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeploymentIsReady'
type DeploymentServiceMock_DeploymentIsReady_Call struct {
	*mock.Call
}

// DeploymentIsReady is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
//   - deploymentName string
func (_e *DeploymentServiceMock_Expecter) DeploymentIsReady(ctx interface{}, projectID interface{}, deploymentName interface{}) *DeploymentServiceMock_DeploymentIsReady_Call {
	return &DeploymentServiceMock_DeploymentIsReady_Call{Call: _e.mock.On("DeploymentIsReady", ctx, projectID, deploymentName)}
}

func (_c *DeploymentServiceMock_DeploymentIsReady_Call) Run(run func(ctx context.Context, projectID string, deploymentName string)) *DeploymentServiceMock_DeploymentIsReady_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *DeploymentServiceMock_DeploymentIsReady_Call) Return(_a0 bool, _a1 error) *DeploymentServiceMock_DeploymentIsReady_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_DeploymentIsReady_Call) RunAndReturn(run func(context.Context, string, string) (bool, error)) *DeploymentServiceMock_DeploymentIsReady_Call {
	_c.Call.Return(run)
	return _c
}

// GetDeployment provides a mock function with given fields: ctx, projectID, _a2
func (_m *DeploymentServiceMock) GetDeployment(ctx context.Context, projectID string, _a2 *v1.AtlasDeployment) (deployment.Deployment, error) {
	ret := _m.Called(ctx, projectID, _a2)

	if len(ret) == 0 {
		panic("no return value specified for GetDeployment")
	}

	var r0 deployment.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1.AtlasDeployment) (deployment.Deployment, error)); ok {
		return rf(ctx, projectID, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *v1.AtlasDeployment) deployment.Deployment); ok {
		r0 = rf(ctx, projectID, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(deployment.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *v1.AtlasDeployment) error); ok {
		r1 = rf(ctx, projectID, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_GetDeployment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDeployment'
type DeploymentServiceMock_GetDeployment_Call struct {
	*mock.Call
}

// GetDeployment is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
//   - _a2 *v1.AtlasDeployment
func (_e *DeploymentServiceMock_Expecter) GetDeployment(ctx interface{}, projectID interface{}, _a2 interface{}) *DeploymentServiceMock_GetDeployment_Call {
	return &DeploymentServiceMock_GetDeployment_Call{Call: _e.mock.On("GetDeployment", ctx, projectID, _a2)}
}

func (_c *DeploymentServiceMock_GetDeployment_Call) Run(run func(ctx context.Context, projectID string, _a2 *v1.AtlasDeployment)) *DeploymentServiceMock_GetDeployment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*v1.AtlasDeployment))
	})
	return _c
}

func (_c *DeploymentServiceMock_GetDeployment_Call) Return(_a0 deployment.Deployment, _a1 error) *DeploymentServiceMock_GetDeployment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_GetDeployment_Call) RunAndReturn(run func(context.Context, string, *v1.AtlasDeployment) (deployment.Deployment, error)) *DeploymentServiceMock_GetDeployment_Call {
	_c.Call.Return(run)
	return _c
}

// ListDeploymentConnections provides a mock function with given fields: ctx, projectID
func (_m *DeploymentServiceMock) ListDeploymentConnections(ctx context.Context, projectID string) ([]deployment.Connection, error) {
	ret := _m.Called(ctx, projectID)

	if len(ret) == 0 {
		panic("no return value specified for ListDeploymentConnections")
	}

	var r0 []deployment.Connection
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]deployment.Connection, error)); ok {
		return rf(ctx, projectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []deployment.Connection); ok {
		r0 = rf(ctx, projectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]deployment.Connection)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, projectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_ListDeploymentConnections_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListDeploymentConnections'
type DeploymentServiceMock_ListDeploymentConnections_Call struct {
	*mock.Call
}

// ListDeploymentConnections is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
func (_e *DeploymentServiceMock_Expecter) ListDeploymentConnections(ctx interface{}, projectID interface{}) *DeploymentServiceMock_ListDeploymentConnections_Call {
	return &DeploymentServiceMock_ListDeploymentConnections_Call{Call: _e.mock.On("ListDeploymentConnections", ctx, projectID)}
}

func (_c *DeploymentServiceMock_ListDeploymentConnections_Call) Run(run func(ctx context.Context, projectID string)) *DeploymentServiceMock_ListDeploymentConnections_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DeploymentServiceMock_ListDeploymentConnections_Call) Return(_a0 []deployment.Connection, _a1 error) *DeploymentServiceMock_ListDeploymentConnections_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_ListDeploymentConnections_Call) RunAndReturn(run func(context.Context, string) ([]deployment.Connection, error)) *DeploymentServiceMock_ListDeploymentConnections_Call {
	_c.Call.Return(run)
	return _c
}

// ListDeploymentNames provides a mock function with given fields: ctx, projectID
func (_m *DeploymentServiceMock) ListDeploymentNames(ctx context.Context, projectID string) ([]string, error) {
	ret := _m.Called(ctx, projectID)

	if len(ret) == 0 {
		panic("no return value specified for ListDeploymentNames")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]string, error)); ok {
		return rf(ctx, projectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, projectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, projectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_ListDeploymentNames_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListDeploymentNames'
type DeploymentServiceMock_ListDeploymentNames_Call struct {
	*mock.Call
}

// ListDeploymentNames is a helper method to define mock.On call
//   - ctx context.Context
//   - projectID string
func (_e *DeploymentServiceMock_Expecter) ListDeploymentNames(ctx interface{}, projectID interface{}) *DeploymentServiceMock_ListDeploymentNames_Call {
	return &DeploymentServiceMock_ListDeploymentNames_Call{Call: _e.mock.On("ListDeploymentNames", ctx, projectID)}
}

func (_c *DeploymentServiceMock_ListDeploymentNames_Call) Run(run func(ctx context.Context, projectID string)) *DeploymentServiceMock_ListDeploymentNames_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *DeploymentServiceMock_ListDeploymentNames_Call) Return(_a0 []string, _a1 error) *DeploymentServiceMock_ListDeploymentNames_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_ListDeploymentNames_Call) RunAndReturn(run func(context.Context, string) ([]string, error)) *DeploymentServiceMock_ListDeploymentNames_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateDeployment provides a mock function with given fields: ctx, _a1
func (_m *DeploymentServiceMock) UpdateDeployment(ctx context.Context, _a1 deployment.Deployment) (deployment.Deployment, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDeployment")
	}

	var r0 deployment.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment) (deployment.Deployment, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment) deployment.Deployment); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(deployment.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, deployment.Deployment) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_UpdateDeployment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateDeployment'
type DeploymentServiceMock_UpdateDeployment_Call struct {
	*mock.Call
}

// UpdateDeployment is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 deployment.Deployment
func (_e *DeploymentServiceMock_Expecter) UpdateDeployment(ctx interface{}, _a1 interface{}) *DeploymentServiceMock_UpdateDeployment_Call {
	return &DeploymentServiceMock_UpdateDeployment_Call{Call: _e.mock.On("UpdateDeployment", ctx, _a1)}
}

func (_c *DeploymentServiceMock_UpdateDeployment_Call) Run(run func(ctx context.Context, _a1 deployment.Deployment)) *DeploymentServiceMock_UpdateDeployment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(deployment.Deployment))
	})
	return _c
}

func (_c *DeploymentServiceMock_UpdateDeployment_Call) Return(_a0 deployment.Deployment, _a1 error) *DeploymentServiceMock_UpdateDeployment_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_UpdateDeployment_Call) RunAndReturn(run func(context.Context, deployment.Deployment) (deployment.Deployment, error)) *DeploymentServiceMock_UpdateDeployment_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateProcessArgs provides a mock function with given fields: ctx, cluster
func (_m *DeploymentServiceMock) UpdateProcessArgs(ctx context.Context, cluster *deployment.Cluster) error {
	ret := _m.Called(ctx, cluster)

	if len(ret) == 0 {
		panic("no return value specified for UpdateProcessArgs")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *deployment.Cluster) error); ok {
		r0 = rf(ctx, cluster)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeploymentServiceMock_UpdateProcessArgs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateProcessArgs'
type DeploymentServiceMock_UpdateProcessArgs_Call struct {
	*mock.Call
}

// UpdateProcessArgs is a helper method to define mock.On call
//   - ctx context.Context
//   - cluster *deployment.Cluster
func (_e *DeploymentServiceMock_Expecter) UpdateProcessArgs(ctx interface{}, cluster interface{}) *DeploymentServiceMock_UpdateProcessArgs_Call {
	return &DeploymentServiceMock_UpdateProcessArgs_Call{Call: _e.mock.On("UpdateProcessArgs", ctx, cluster)}
}

func (_c *DeploymentServiceMock_UpdateProcessArgs_Call) Run(run func(ctx context.Context, cluster *deployment.Cluster)) *DeploymentServiceMock_UpdateProcessArgs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*deployment.Cluster))
	})
	return _c
}

func (_c *DeploymentServiceMock_UpdateProcessArgs_Call) Return(_a0 error) *DeploymentServiceMock_UpdateProcessArgs_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DeploymentServiceMock_UpdateProcessArgs_Call) RunAndReturn(run func(context.Context, *deployment.Cluster) error) *DeploymentServiceMock_UpdateProcessArgs_Call {
	_c.Call.Return(run)
	return _c
}

// UpgradeToDedicated provides a mock function with given fields: ctx, currentDeployment, targetDeployment
func (_m *DeploymentServiceMock) UpgradeToDedicated(ctx context.Context, currentDeployment deployment.Deployment, targetDeployment deployment.Deployment) (deployment.Deployment, error) {
	ret := _m.Called(ctx, currentDeployment, targetDeployment)

	if len(ret) == 0 {
		panic("no return value specified for UpgradeToDedicated")
	}

	var r0 deployment.Deployment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment, deployment.Deployment) (deployment.Deployment, error)); ok {
		return rf(ctx, currentDeployment, targetDeployment)
	}
	if rf, ok := ret.Get(0).(func(context.Context, deployment.Deployment, deployment.Deployment) deployment.Deployment); ok {
		r0 = rf(ctx, currentDeployment, targetDeployment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(deployment.Deployment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, deployment.Deployment, deployment.Deployment) error); ok {
		r1 = rf(ctx, currentDeployment, targetDeployment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeploymentServiceMock_UpgradeToDedicated_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpgradeToDedicated'
type DeploymentServiceMock_UpgradeToDedicated_Call struct {
	*mock.Call
}

// UpgradeToDedicated is a helper method to define mock.On call
//   - ctx context.Context
//   - currentDeployment deployment.Deployment
//   - targetDeployment deployment.Deployment
func (_e *DeploymentServiceMock_Expecter) UpgradeToDedicated(ctx interface{}, currentDeployment interface{}, targetDeployment interface{}) *DeploymentServiceMock_UpgradeToDedicated_Call {
	return &DeploymentServiceMock_UpgradeToDedicated_Call{Call: _e.mock.On("UpgradeToDedicated", ctx, currentDeployment, targetDeployment)}
}

func (_c *DeploymentServiceMock_UpgradeToDedicated_Call) Run(run func(ctx context.Context, currentDeployment deployment.Deployment, targetDeployment deployment.Deployment)) *DeploymentServiceMock_UpgradeToDedicated_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(deployment.Deployment), args[2].(deployment.Deployment))
	})
	return _c
}

func (_c *DeploymentServiceMock_UpgradeToDedicated_Call) Return(_a0 deployment.Deployment, _a1 error) *DeploymentServiceMock_UpgradeToDedicated_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DeploymentServiceMock_UpgradeToDedicated_Call) RunAndReturn(run func(context.Context, deployment.Deployment, deployment.Deployment) (deployment.Deployment, error)) *DeploymentServiceMock_UpgradeToDedicated_Call {
	_c.Call.Return(run)
	return _c
}

// NewDeploymentServiceMock creates a new instance of DeploymentServiceMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDeploymentServiceMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *DeploymentServiceMock {
	mock := &DeploymentServiceMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

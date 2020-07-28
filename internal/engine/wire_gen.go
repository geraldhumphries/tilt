// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package engine

import (
	"context"

	"github.com/google/wire"
	"github.com/tilt-dev/wmclient/pkg/dirs"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/tilt-dev/tilt/internal/analytics"
	"github.com/tilt-dev/tilt/internal/build"
	"github.com/tilt-dev/tilt/internal/containerupdate"
	"github.com/tilt-dev/tilt/internal/docker"
	"github.com/tilt-dev/tilt/internal/dockercompose"
	"github.com/tilt-dev/tilt/internal/dockerfile"
	"github.com/tilt-dev/tilt/internal/engine/buildcontrol"
	"github.com/tilt-dev/tilt/internal/k8s"
	"github.com/tilt-dev/tilt/internal/synclet"
	"github.com/tilt-dev/tilt/internal/synclet/sidecar"
	"github.com/tilt-dev/tilt/internal/tracer"
)

// Injectors from wire.go:

func provideBuildAndDeployer(ctx context.Context, docker2 docker.Client, kClient k8s.Client, dir *dirs.WindmillDir, env k8s.Env, updateMode buildcontrol.UpdateModeFlag, sCli *synclet.TestSyncletClient, dcc dockercompose.DockerComposeClient, clock build.Clock, kp KINDLoader, analytics2 *analytics.TiltAnalytics) (BuildAndDeployer, error) {
	dockerContainerUpdater := containerupdate.NewDockerUpdater(docker2)
	syncletClient, err := synclet.FakeGRPCWrapper(ctx, sCli)
	if err != nil {
		return nil, err
	}
	syncletManager := containerupdate.NewSyncletManagerForTests(kClient, syncletClient, sCli)
	syncletUpdater := containerupdate.NewSyncletUpdater(syncletManager)
	execUpdater := containerupdate.NewExecUpdater(kClient)
	runtime := k8s.ProvideContainerRuntime(ctx, kClient)
	buildcontrolUpdateMode, err := buildcontrol.ProvideUpdateMode(updateMode, env, runtime)
	if err != nil {
		return nil, err
	}
	liveUpdateBuildAndDeployer := NewLiveUpdateBuildAndDeployer(dockerContainerUpdater, syncletUpdater, execUpdater, buildcontrolUpdateMode, env, runtime, clock)
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(docker2, labels)
	dockerBuilder := build.DefaultDockerBuilder(dockerImageBuilder)
	execCustomBuilder := build.NewExecCustomBuilder(docker2, clock)
	syncletImageRef, err := sidecar.ProvideSyncletImageRef(ctx)
	if err != nil {
		return nil, err
	}
	syncletContainer := sidecar.ProvideSyncletContainer(syncletImageRef)
	imageBuildAndDeployer := NewImageBuildAndDeployer(dockerBuilder, execCustomBuilder, kClient, env, analytics2, buildcontrolUpdateMode, clock, runtime, kp, syncletContainer)
	engineImageBuilder := NewImageBuilder(dockerBuilder, execCustomBuilder, buildcontrolUpdateMode)
	dockerComposeBuildAndDeployer := NewDockerComposeBuildAndDeployer(dcc, docker2, engineImageBuilder, clock)
	localTargetBuildAndDeployer := NewLocalTargetBuildAndDeployer(clock)
	buildOrder := DefaultBuildOrder(liveUpdateBuildAndDeployer, imageBuildAndDeployer, dockerComposeBuildAndDeployer, localTargetBuildAndDeployer, buildcontrolUpdateMode, env, runtime)
	spanProcessor := _wireSpanProcessorValue
	traceTracer, err := tracer.InitOpenTelemetry(ctx, spanProcessor)
	if err != nil {
		return nil, err
	}
	compositeBuildAndDeployer := NewCompositeBuildAndDeployer(buildOrder, traceTracer)
	return compositeBuildAndDeployer, nil
}

var (
	_wireLabelsValue        = dockerfile.Labels{}
	_wireSpanProcessorValue = (trace.SpanProcessor)(nil)
)

func provideImageBuildAndDeployer(ctx context.Context, docker2 docker.Client, kClient k8s.Client, env k8s.Env, dir *dirs.WindmillDir, clock build.Clock, kp KINDLoader, analytics2 *analytics.TiltAnalytics) (*ImageBuildAndDeployer, error) {
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(docker2, labels)
	dockerBuilder := build.DefaultDockerBuilder(dockerImageBuilder)
	execCustomBuilder := build.NewExecCustomBuilder(docker2, clock)
	updateModeFlag := _wireUpdateModeFlagValue
	runtime := k8s.ProvideContainerRuntime(ctx, kClient)
	updateMode, err := buildcontrol.ProvideUpdateMode(updateModeFlag, env, runtime)
	if err != nil {
		return nil, err
	}
	syncletImageRef, err := sidecar.ProvideSyncletImageRef(ctx)
	if err != nil {
		return nil, err
	}
	syncletContainer := sidecar.ProvideSyncletContainer(syncletImageRef)
	imageBuildAndDeployer := NewImageBuildAndDeployer(dockerBuilder, execCustomBuilder, kClient, env, analytics2, updateMode, clock, runtime, kp, syncletContainer)
	return imageBuildAndDeployer, nil
}

var (
	_wireUpdateModeFlagValue = buildcontrol.UpdateModeFlag(buildcontrol.UpdateModeAuto)
)

func provideDockerComposeBuildAndDeployer(ctx context.Context, dcCli dockercompose.DockerComposeClient, dCli docker.Client, dir *dirs.WindmillDir) (*DockerComposeBuildAndDeployer, error) {
	labels := _wireLabelsValue
	dockerImageBuilder := build.NewDockerImageBuilder(dCli, labels)
	dockerBuilder := build.DefaultDockerBuilder(dockerImageBuilder)
	clock := build.ProvideClock()
	execCustomBuilder := build.NewExecCustomBuilder(dCli, clock)
	updateModeFlag := _wireBuildcontrolUpdateModeFlagValue
	env := _wireEnvValue
	clientConfig := k8s.ProvideClientConfig()
	restConfigOrError := k8s.ProvideRESTConfig(clientConfig)
	clientsetOrError := k8s.ProvideClientset(restConfigOrError)
	portForwardClient := k8s.ProvidePortForwardClient(restConfigOrError, clientsetOrError)
	namespace := k8s.ProvideConfigNamespace(clientConfig)
	config, err := k8s.ProvideKubeConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	kubeContext, err := k8s.ProvideKubeContext(config)
	if err != nil {
		return nil, err
	}
	int2 := provideKubectlLogLevelInfo()
	kubectlRunner := k8s.ProvideKubectlRunner(kubeContext, int2)
	minikubeClient := k8s.ProvideMinikubeClient(kubeContext)
	client := k8s.ProvideK8sClient(ctx, env, restConfigOrError, clientsetOrError, portForwardClient, namespace, kubectlRunner, minikubeClient, clientConfig)
	runtime := k8s.ProvideContainerRuntime(ctx, client)
	updateMode, err := buildcontrol.ProvideUpdateMode(updateModeFlag, env, runtime)
	if err != nil {
		return nil, err
	}
	engineImageBuilder := NewImageBuilder(dockerBuilder, execCustomBuilder, updateMode)
	dockerComposeBuildAndDeployer := NewDockerComposeBuildAndDeployer(dcCli, dCli, engineImageBuilder, clock)
	return dockerComposeBuildAndDeployer, nil
}

var (
	_wireBuildcontrolUpdateModeFlagValue = buildcontrol.UpdateModeFlag(buildcontrol.UpdateModeAuto)
	_wireEnvValue                        = k8s.Env(k8s.EnvNone)
)

// wire.go:

var DeployerBaseWireSet = wire.NewSet(wire.Value(dockerfile.Labels{}), wire.Value(UpperReducer), sidecar.WireSet, k8s.ProvideMinikubeClient, build.DefaultDockerBuilder, build.NewDockerImageBuilder, build.NewExecCustomBuilder, wire.Bind(new(build.CustomBuilder), new(*build.ExecCustomBuilder)), NewLocalTargetBuildAndDeployer,
	NewImageBuildAndDeployer, containerupdate.NewDockerUpdater, containerupdate.NewSyncletUpdater, containerupdate.NewExecUpdater, NewLiveUpdateBuildAndDeployer,
	NewDockerComposeBuildAndDeployer,
	NewImageBuilder,
	DefaultBuildOrder, tracer.InitOpenTelemetry, wire.Bind(new(BuildAndDeployer), new(*CompositeBuildAndDeployer)), NewCompositeBuildAndDeployer, buildcontrol.ProvideUpdateMode,
)

var DeployerWireSetTest = wire.NewSet(
	DeployerBaseWireSet, containerupdate.NewSyncletManagerForTests, wire.InterfaceValue(new(trace.SpanProcessor), (trace.SpanProcessor)(nil)), synclet.FakeGRPCWrapper,
)

var DeployerWireSet = wire.NewSet(
	DeployerBaseWireSet, containerupdate.NewSyncletManager,
)

func provideKubectlLogLevelInfo() k8s.KubectlLogLevel {
	return k8s.KubectlLogLevel(0)
}

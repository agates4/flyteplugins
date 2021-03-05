package k8s

import (
	"context"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

//go:generate mockery -all -case=underscore

// PluginEntry is a structure that is used to indicate to the system a K8s plugin
type PluginEntry struct {
	// ID/Name of the plugin. This will be used to identify this plugin and has to be unique in the entire system
	// All functions like enabling and disabling a plugin use this ID
	ID pluginsCore.TaskType
	// A list of all the task types for which this plugin is applicable.
	RegisteredTaskTypes []pluginsCore.TaskType
	// An instance of the kubernetes resource this plugin is responsible for, for example v1.Pod{}
	ResourceToWatch runtime.Object
	// An instance of the plugin
	Plugin Plugin
	// Boolean that indicates if this plugin can be used as the default for unknown task types. There can only be
	// one default in the system
	IsDefault bool
	// A list of all task types for which this plugin should be default handler when multiple registered plugins
	// support the same task type. This must be a subset of RegisteredTaskTypes and at most one default per task type
	// is supported.
	DefaultForTaskTypes []pluginsCore.TaskType
}

// A proxy object for k8s resource
type Resource interface {
	runtime.Object
	metav1.Object
	schema.ObjectKind
}

// Special context passed in to plugins when checking task phase
type PluginContext interface {
	// Returns a TaskReader, to retrieve task details
	TaskReader() pluginsCore.TaskReader

	// Returns an input reader to retrieve input data
	InputReader() io.InputReader

	// Provides an output sync of type io.OutputWriter
	OutputWriter() io.OutputWriter

	// Returns a handle to the currently configured storage backend that can be used to communicate with the tasks or write metadata
	DataStore() *storage.DataStore

	// Returns the max allowed dataset size that the outputwriter will accept
	MaxDatasetSizeBytes() int64
}

// Defines a simplified interface to author plugins for k8s resources.
type Plugin interface {
	// Defines a func to create a query object (typically just object and type meta portions) that's used to query k8s
	// resources.
	BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (Resource, error)

	// Defines a func to create the full resource object that will be posted to k8s.
	BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (Resource, error)

	// Analyses the k8s resource and reports the status as TaskPhase. This call is expected to be relatively fast,
	// any operations that might take a long time (limits are configured system-wide) should be offloaded to the
	// background.
	GetTaskPhase(ctx context.Context, pluginContext PluginContext, resource Resource) (pluginsCore.PhaseInfo, error)
}

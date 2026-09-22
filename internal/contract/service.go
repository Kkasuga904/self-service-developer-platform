package contract

const (
	APIVersion = "platform.example.io/v1alpha1"
	Kind       = "Service"
)

type Service struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

type Metadata struct {
	Name string `yaml:"name" json:"name"`
}

type Spec struct {
	Owner         string        `yaml:"owner" json:"owner"`
	Image         string        `yaml:"image" json:"image"`
	Port          int           `yaml:"port" json:"port"`
	Resources     Resources     `yaml:"resources" json:"resources"`
	Availability  Availability  `yaml:"availability" json:"availability"`
	Observability Observability `yaml:"observability" json:"observability"`
	Health        Health        `yaml:"health,omitempty" json:"health,omitempty"`
	Autoscaling   *Autoscaling  `yaml:"autoscaling,omitempty" json:"autoscaling,omitempty"`
}

type Resources struct {
	Size string `yaml:"size" json:"size"`
}

type Availability struct {
	Replicas int `yaml:"replicas" json:"replicas"`
}

type Observability struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type Health struct {
	LivenessPath  string `yaml:"livenessPath,omitempty" json:"livenessPath,omitempty"`
	ReadinessPath string `yaml:"readinessPath,omitempty" json:"readinessPath,omitempty"`
}

type Autoscaling struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`
	MinReplicas int  `yaml:"minReplicas,omitempty" json:"minReplicas,omitempty"`
	MaxReplicas int  `yaml:"maxReplicas,omitempty" json:"maxReplicas,omitempty"`
	TargetCPU   int  `yaml:"targetCPUUtilization,omitempty" json:"targetCPUUtilization,omitempty"`
}

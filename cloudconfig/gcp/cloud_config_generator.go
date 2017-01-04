package gcp

import yaml "gopkg.in/yaml.v2"

type CloudConfigGenerator struct{}

type CloudConfigInput struct {
	AZs                 []string
	Tags                []string
	NetworkName         string
	SubnetworkName      string
	ConcourseTargetPool string
	CFBackends          CFBackends
}

type CFBackends struct {
	Router   string
	SSHProxy string
}

type VMExtension struct {
	Name            string                     `yaml:"name"`
	CloudProperties VMExtensionCloudProperties `yaml:"cloud_properties"`
}

type VMExtensionCloudProperties struct {
	RootDiskSizeGB      int      `yaml:"root_disk_size_gb,omitempty"`
	RootDiskType        string   `yaml:"root_disk_type,omitempty"`
	TargetPool          string   `yaml:"target_pool,omitempty"`
	EphemeralExternalIP *bool    `yaml:"ephemeral_external_ip,omitempty"`
	BackendService      string   `yaml:"backend_service,omitempty"`
	Tags                []string `yaml:"tags,omitempty"`
}

type CloudConfig struct {
	AZs          []AZ          `yaml:"azs,omitempty"`
	Networks     []Network     `yaml:"networks,omitempty"`
	VMTypes      interface{}   `yaml:"vm_types,omitempty"`
	DiskTypes    interface{}   `yaml:"disk_types,omitempty"`
	Compilation  interface{}   `yaml:"compilation,omitempty"`
	VMExtensions []VMExtension `yaml:"vm_extensions,omitempty"`
}

var unmarshal func([]byte, interface{}) error = yaml.Unmarshal

func NewCloudConfigGenerator() CloudConfigGenerator {
	return CloudConfigGenerator{}
}

func (c CloudConfigGenerator) Generate(input CloudConfigInput) (CloudConfig, error) {
	var cloudConfig CloudConfig
	if err := unmarshal([]byte(cloudConfigTemplate), &cloudConfig); err != nil {
		return CloudConfig{}, err
	}

	cloudConfig = c.generateAZs(input, cloudConfig)

	cloudConfig, err := c.generateNetworks(input, cloudConfig)
	if err != nil {
		return CloudConfig{}, err
	}

	if input.ConcourseTargetPool != "" {
		cloudConfig.VMExtensions = append(cloudConfig.VMExtensions, VMExtension{
			Name: "lb",
			CloudProperties: VMExtensionCloudProperties{
				TargetPool: input.ConcourseTargetPool,
			},
		})
	}

	if input.CFBackends.Router != "" {
		cloudConfig.VMExtensions = append(cloudConfig.VMExtensions, VMExtension{
			Name: "router-lb",
			CloudProperties: VMExtensionCloudProperties{
				BackendService: input.CFBackends.Router,
				Tags:           []string{input.CFBackends.Router},
			},
		})
	}

	if input.CFBackends.SSHProxy != "" {
		cloudConfig.VMExtensions = append(cloudConfig.VMExtensions, VMExtension{
			Name: "ssh-proxy-lb",
			CloudProperties: VMExtensionCloudProperties{
				TargetPool: input.CFBackends.SSHProxy,
				Tags:       []string{input.CFBackends.SSHProxy},
			},
		})
	}

	return cloudConfig, nil
}

func (CloudConfigGenerator) generateAZs(input CloudConfigInput, cloudConfig CloudConfig) CloudConfig {
	azsGenerator := NewAZsGenerator(input.AZs...)
	cloudConfig.AZs = azsGenerator.Generate()
	return cloudConfig
}

func (CloudConfigGenerator) generateNetworks(input CloudConfigInput, cloudConfig CloudConfig) (CloudConfig, error) {
	azs := []string{}
	for _, az := range cloudConfig.AZs {
		azs = append(azs, az.Name)
	}

	networksGenerator := NewNetworksGenerator(input.NetworkName, input.SubnetworkName, input.Tags, azs)

	var err error
	cloudConfig.Networks, err = networksGenerator.Generate()
	if err != nil {
		return CloudConfig{}, err
	}

	return cloudConfig, nil
}

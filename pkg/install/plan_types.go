package install

import "fmt"

// SSHDetails is an interface to allow to SSH into nodes
type SSHDetails interface {
	GetSSHAddress()
	GetSSHPort()
	GetSSHKeyPath()
	GetSSHUsername()
}

// NetworkConfig describes the cluster's networking configuration
type NetworkConfig struct {
	Type             string
	PodCIDRBlock     string `yaml:"pod_cidr_block"`
	ServiceCIDRBlock string `yaml:"service_cidr_block"`
	PolicyEnabled    bool   `yaml:"policy_enabled"`
	UpdateHostsFiles bool   `yaml:"update_hosts_files"`
}

// CertsConfig describes the cluster's trust and certificate configuration
type CertsConfig struct {
	Expiry string
}

// SSHConfig describes the cluster's SSH configuration for accessing nodes
type SSHConfig struct {
	User string
	Key  string `yaml:"ssh_key"`
	Port int    `yaml:"ssh_port"`
}

// Cluster describes a Kubernetes cluster
type Cluster struct {
	Name                     string
	AdminPassword            string `yaml:"admin_password"`
	AllowPackageInstallation bool   `yaml:"allow_package_installation"`
	Networking               NetworkConfig
	Certificates             CertsConfig
	SSH                      SSHConfig
}

// A Node is a compute unit, virtual or physical, that is part of the cluster
type Node struct {
	Host       string
	IP         string
	InternalIP string
}

// A NodeGroup is a collection of nodes
type NodeGroup struct {
	ExpectedCount int `yaml:"expected_count"`
	Nodes         []Node
}

// An OptionalNodeGroup is a collection of nodes that can be empty
type OptionalNodeGroup NodeGroup

// MasterNodeGroup is the collection of master nodes
type MasterNodeGroup struct {
	ExpectedCount         int    `yaml:"expected_count"`
	LoadBalancedFQDN      string `yaml:"load_balanced_fqdn"`
	LoadBalancedShortName string `yaml:"load_balanced_short_name"`
	Nodes                 []Node
}

// DockerRegistry details for docker registry, either confgiured by the cli or customer provided
type DockerRegistry struct {
	SetupInternal bool `yaml:"setup_internal"`
	Address       string
	Port          int
	CAPath        string `yaml:"CA"`
}

// Plan is the installation plan that the user intends to execute
type Plan struct {
	Cluster        Cluster
	DockerRegistry DockerRegistry `yaml:"docker_registry"`
	Etcd           NodeGroup
	Master         MasterNodeGroup
	Worker         NodeGroup
	Ingress        OptionalNodeGroup
}

type SSHConnections struct {
	SSHConfig *SSHConfig
	Nodes     []Node
	Retries   uint
}

type SSHConnection struct {
	SSHConfig *SSHConfig
	Node      *Node
	Retries   uint
}

// GetSSHConnection returns the SSHConnection struct containing the node and SSHConfig details
func (p *Plan) GetSSHConnection(host string) (*SSHConnection, error) {
	nodes := []Node{}
	nodes = append(nodes, p.Etcd.Nodes...)
	nodes = append(nodes, p.Master.Nodes...)
	nodes = append(nodes, p.Worker.Nodes...)
	if p.Ingress.Nodes != nil {
		nodes = append(nodes, p.Ingress.Nodes...)
	}
	// try to find the node with the provided hostname
	var foundNode *Node
	for _, node := range nodes {
		if node.Host == host {
			foundNode = &node
			break
		}
	}

	if foundNode == nil {
		return nil, fmt.Errorf("node %q not found in the plan", host)
	}

	return &SSHConnection{&p.Cluster.SSH, foundNode, 1}, nil
}

func (ssh *SSHConnection) GetSSHAddress() string {
	return ssh.Node.IP
}

func (ssh *SSHConnection) GetSSHPort() int {
	return ssh.SSHConfig.Port
}

func (ssh *SSHConnection) GetSSHKeyPath() string {
	return ssh.SSHConfig.Key
}

func (ssh *SSHConnection) GetSSHUsername() string {
	return ssh.SSHConfig.User
}

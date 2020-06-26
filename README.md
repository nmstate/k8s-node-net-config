# kubernetes-nmstate

[keɪ ɛn ɛm steɪt] Declarative node network configuration driven through Kubernetes API.

# How it works

We use [nmstate](https://nmstate.io/) to perform state driven network
configuration on cluster nodes and to report back their current state.
Both the configuration and reporting is controlled via Kubernetes objects.

```yaml
apiVersion: nmstate.io/v1alpha1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: br1-eth0
spec:
  desiredState:
    interfaces:
    - name: br1
      type: linux-bridge
      state: up
      ipv4:
        dhcp: true
        enabled: true
      bridge:
        port:
        - name: eth0
```

The only external dependency is NetworkManager running on nodes. See more
details in
[Compatibility documentation](CONTRIBUTING.md#networkmanager-compatibility).

# Deployment and Usage

You can choose to deploy this operator on a
[local virtualized cluster](docs/deployment-local-cluster.md) or on your
[arbitrary cluster](docs/deployment-arbitrary-cluster.md).

Following comprehensive 101 series is the best place to start learning about all
the features:

1. [Reporting](docs/user-guide-101-reporting.md) -
   observe the current state of network on cluster nodes.
2. [Configuring](docs/user-guide-102-configuration.md) -
   configure networks and observe the progress.
3. [Troubleshooting](docs/user-guide-103-troubleshooting.md) -
   see what's wrong if a configuration fails.

These example manifests should serve as reference on how to configure various
configuration options:

- [Linux bonding](docs/examples/bond.yaml)
- [Linux bonding with VLAN](docs/examples/bond-vlan.yaml)
- [Linux bridge](docs/examples/linux-bridge.yaml)
- [VLAN](docs/examples/vlan.yaml)
- [Ethernet](docs/examples/ethernet.yaml)
- [Open vSwitch bridge](docs/examples/ovs-bridge.yaml)
- [Open vSwitch bridge interface](docs/examples/ovs-bridge-iface.yaml)
- [Static IP](docs/examples/static-ip.yaml)
- [DHCP](docs/examples/dhcp.yaml)
- [Routes](docs/examples/route.yaml)
- [DNS](docs/examples/dns.yaml)
- [Workers selector](docs/examples/worker-selector.yaml)

# The "Why"

With hybrid clouds, node-networking setup is becoming even more challenging.
Different payloads have different networking requirements, and not everything
can be satisfied as overlays on top of the main interface of the node (e.g.
SR-IOV, L2, other L2).
The [Container Network Interface](https://github.com/containernetworking/cni)
(CNI) standard enables different
solutions for connecting networks on the node with pods. Some of them are
[part of the standard](https://github.com/containernetworking/plugins), and there are
others that extend support for [Open vSwitch bridges](https://github.com/kubevirt/ovs-cni),
[SR-IOV](https://github.com/hustcat/sriov-cni), and more...

However, in all of these cases, the node must have the networks setup before the
pod is scheduled. Setting up the networks in a dynamic and heterogenous cluster,
with dynamic networking requirements, is a challenge by itself - and this is
what this project is addressing.

# Development and Contributing

Contributions are welcome! Find details about the project's design and
development workflow in the [developer guide](CONTRIBUTING.md).


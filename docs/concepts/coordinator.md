# Coordinator

**English** | [**简体中文**](coordinator-zh_CN.md)

Spiderpool incorporates a CNI meta-plugin called `coordinator` that works after the Main CNI is invoked. It mainly offers the following features:

- Resolve the problem of underlay Pods unable to access ClusterIP
- Coordinate the routing for Pods with multiple NICs, ensuring consistent packet paths
- Detect IP conflicts within Pods
- Check the reachability of Pod gateways
- Support fixed Mac address prefixes for Pods

Let's delve into how coordinator implements these features.

> You can configure `coordinator` by specifying all the relevant fields in `SpinderMultusConfig` if a NetworkAttachmentDefinition CR is created via `SpinderMultusConfig CR`. For more information, please refer to [SpinderMultusConfig](../reference/crd-spidermultusconfig.md).
>
> `Spidercoordinators CR` serves as the global default configuration (all fields) for `coordinator`. However, this configuration has a lower priority compared to the settings in the NetworkAttachmentDefinition CR. In cases where no configuration is provided in the NetworkAttachmentDefinition CR, the values from `Spidercoordinators CR` serve as the defaults. For detailed information, please refer to [Spidercoordinator](../reference/crd-spidercoordinator.md).

## Resolve the problem of underlay Pods unable to access ClusterIP(beta)

When using underlay CNIs like Macvlan, IPvlan, SR-IOV, and others, a common challenge arises where underlay pods are unable to access ClusterIP. This occurs because accessing ClusterIP from underlay pods requires routing through the gateway on the switch. However, in many instances, the gateway is not configured with the proper routes to reach the ClusterIP, leading to restricted access.

For more information about the Underlay Pod not being able to access the ClusterIP, please refer to [Underlay CNI Access Service](../usage/underlay_cni_service.md)

## Detect Pod IP conflicts(alpha)

IP conflicts are unacceptable for underlay networks, which can cause serious problems. When creating a pod, we can use the `coordinator` to detect whether the IP of the pod conflicts, and support both IPv4 and IPv6 addresses. By sending an ARP or NDP probe message,
If the MAC address of the reply packet is not the pod itself, we consider the IP to be conflicting and reject the creation of the pod with conflicting IP addresses:

```yaml
apiVersion: spiderpool.spidernet.io/v2beta1
kind: SpiderMultusConfig
metadata:
  name: detect-ip
  namespace: default
spec:
  cniType: macvlan
  macvlan:
    master: ["eth0"]
  coordinator:
    detectIPConflict: true    # Enable detectIPConflict
```

## Detect Pod gateway reachability(alpha)

Under the underlay network, pod access to the outside needs to be forwarded through the gateway. If the gateway is unreachable, then the pod is actually lost. Sometimes we want to create a pod with a gateway reachable. We can use the 'coordinator' to check if the pod's gateway is reachable.
Gateway addresses for IPv4 and IPv6 can be detected. We send an ICMP packet to check whether the gateway address is reachable. If the gateway is unreachable, pods will be prevented from creating:

We can configure it via Spidermultusconfig:

```yaml
apiVersion: spiderpool.spidernet.io/v2beta1
kind: SpiderMultusConfig
metadata:
  name: detect-gateway
  namespace: default
spec:
  cniType: macvlan
  macvlan:
    master: ["eth0"]
  enableCoordinator: true
  coordinator:
    detectGateway: true    # Enable detectGateway
```

> Note: There are some switches that are not allowed to be probed by arp, otherwise an alarm will be issued, in this case, we need to set detectGateway to false


## Fix MAC address prefix for Pods(alpha)

Some traditional applications may require a fixed MAC address or IP address to couple the behavior of the application. For example, the License Server may need to apply a fixed Mac address 
or IP address to issue a license for the app. If the MAC address of a pod changes, the issued license may be invalid. Therefore, you need to fix the MAC address of the pod. Spiderpool can fix 
the MAC address of the application through `coordinator`, and the fixed rule is to configure the MAC address prefix (2 bytes) + convert the IP of the pod (4 bytes).

Note:

> currently supports updating  Macvlan and SR-IOV as pods for CNI. In IPVlan L2 mode, the MAC addresses of the primary interface and the sub-interface are the same and cannot be modified.
> 
> The fixed rule is to configure the MAC address prefix (2 bytes) + the IP of the converted pod (4 bytes). An IPv4 address is 4 bytes long and can be fully converted to 2 hexadecimal numbers. For IPv6 addresses, only the last 4 bytes are taken.

We can configure it via Spidermultusconfig:

```yaml
apiVersion: spiderpool.spidernet.io/v2beta1
kind: SpiderMultusConfig
metadata:
  name: overwrite-mac
  namespace: default
spec:
  cniType: macvlan
  macvlan:
    master: ["eth0"]
  enableCoordinator: true
  coordinator:
    podMACPrefix: "0a:1b"    # Enable detectGateway
```

You can check if the MAC address prefix of the Pod starts with "0a:1b" after a Pod is created.

## Known issues

- Underlay mode: TCP communication between underlay Pods and overlay Pods (Calico or Cilium) fails

    This issue arises from inconsistent packet routing paths. Request packets are matched with the routing on the source Pod side and forwarded through veth0 to the host side. And then the packets are further forwarded to the target Pod. The target Pod perceives the source IP of the packet as the underlay IP of the source Pod, allowing it to bypass the source Pod's host and directly route through the underlay network.
    However, on the host, this is considered an invalid packet (as it receives unexpected TCP SYN-ACK packets that are conntrack table invalid), explicitly dropping it using an iptables rule in kube-proxy. Switching the kube-proxy mode to ipvs can address this issue. This issue is expected to be fixed in K8s 1.29.
    if the sysctl `nf_conntrack_tcp_be_liberal` is set to 1, kube-proxy will not deliver the DROP rule.

- Overlay mode: with Cilium as the default CNI and multiple NICs for the Pod, the underlay interface of the Pod cannot communicate with the node.

    Macvlan interfaces do not allow direct communication between parent and child interfaces in bridge mode in most cases. To facilitate communication between the underlay IP of the Pod and the node, we rely on the default CNI to create Veth devices. However, Cilium restricts the forwarding of IPs from non-Cilium subnets through these Veth devices.
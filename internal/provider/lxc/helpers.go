package lxc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/iolave/go-proxmox/pkg/pve"
)

func getVMID(c *pve.PVE, data types.Int64) (int, error) {

	if data.IsNull() || data.IsUnknown() {
		vmid, err := c.Cluster.GetRandomVMID()
		if err != nil {
			return 0, err
		}
		return vmid, nil
	}

	return int(data.ValueInt64()), nil
}

func formatSSHPublicKey(keys []types.String) string {
	if len(keys) == 0 {
		return ""
	}

	result := ""
	for _, pub := range keys {
		result = fmt.Sprintf("%s\n%s", result, pub.ValueString())
	}
	return result
}

func newLXCFeaturesResourceModel(ctx context.Context, obj types.Object) LXCFeaturesResourceModel {
	feats := LXCFeaturesResourceModel{}
	feats.LoadFromObject(ctx, obj)
	return feats
}

func newPVELXCNets(ctx context.Context, objs []types.Object) []pve.LxcNet {
	nets := []pve.LxcNet{}
	for _, obj := range objs {
		net := LXCNetResourceModel{}
		net.LoadFromObject(ctx, obj)
		nets = append(nets, net.ToPVELXCNet())
	}
	return nets
}

func updateLXCStatus(
	ctx context.Context,
	c *pve.PVE,
	node string,
	vmid int,
	desiredStatus string,
) error {
	retries := 5
	try := 0
	for true {
		try++
		time.Sleep(time.Second * 8)

		remoteStatus, err := c.LXC.GetStatus(
			node,
			vmid,
		)
		if err != nil {
			return err
		}

		if desiredStatus == remoteStatus.Status {
			break
		}

		tflog.Debug(ctx, "debug_lxc_status", map[string]any{"vmid": vmid, "current_status": remoteStatus.Status, "desired_status": desiredStatus})
		switch desiredStatus {
		case string(pve.LXC_STATUS_RUNNING):
			if remoteStatus.Status != string(pve.LXC_STATUS_STOPPED) {
				continue
			}
			_, err = c.LXC.Start(pve.LXCStartRequest{Node: node, ID: vmid})
			break

		case string(pve.LXC_STATUS_STOPPED):
			if remoteStatus.Status != string(pve.LXC_STATUS_RUNNING) {
				continue
			}
			_, err = c.LXC.Stop(pve.LXCStopRequest{Node: node, ID: vmid})
			break
		default:
			err = fmt.Errorf("unexpected status value, got %s", desiredStatus)
			break
		}

		if try <= retries {
			continue
		} else {
			return err
		}
	}

	return nil
}

func computeLXCNetIPs(
	ctx context.Context,
	c *pve.PVE,
	node string,
	vmid int,
	nets []types.Object,
) ([]types.Object, error) {
	ipRetries := 3
	ipSleepTime := time.Second * 15
	for i := 0; i < ipRetries; i++ {
		time.Sleep(ipSleepTime)

		// retrieve lxc interfaces of running lxc
		// and store them in the map below for easy
		// access through the iface name.
		ifacesMap := map[string]pve.GetLxcInterfaceResponse{}
		ifaces, err := c.LXC.GetInterfaces(node, vmid)
		if err != nil {
			continue
		}
		for _, iface := range ifaces {
			ifacesMap[iface.Name] = iface
		}

		// Read configured networks
		computedNets := map[string]LXCNetResourceModel{}
		for _, obj := range nets {
			net := LXCNetResourceModel{}
			net.LoadFromObject(ctx, obj)

			if ifacesMap[net.Name].IPv4 == "" {
				break
			}

			ip := ifacesMap[net.Name].IPv4
			net.ComputedIP = &ip
			computedNets[net.Name] = net
		}

		if len(nets) != len(computedNets) {
			err := errors.New("Unable to compute all ifaces ips")
			return nil, err
		}

		computedTFNets := []types.Object{}
		for _, netPosObj := range nets {
			netPosModel := LXCNetResourceModel{}
			netPosModel.LoadFromObject(ctx, netPosObj)
			net := computedNets[netPosModel.Name]

			computedTFNets = append(computedTFNets, net.ToObject())
		}
		return computedTFNets, nil
	}

	return nil, fmt.Errorf("Unable to compute all ifaces ips after %d retries", ipRetries)
}

func newLXCNetsResourceModel(ctx context.Context, obj []types.Object) []LXCNetResourceModel {
	nets := []LXCNetResourceModel{}
	for _, netObj := range obj {
		net := LXCNetResourceModel{}
		net.LoadFromObject(ctx, netObj)
		nets = append(nets, net)
	}
	return nets
}

func runLXCCommands(
	ctx context.Context,
	c *pve.PVE,
	vmid int,
	cmds []types.String,
) error {
	for _, cmd := range cmds {
		var execId string
		var execErr error
		cmdstr := cmd.ValueString()

		for retry := 0; retry < 3; retry++ {
			tflog.Info(ctx, "executing cmd", map[string]any{"cmd": cmdstr})
			execId, execErr = c.LXC.ExecAsync(vmid, "bash", cmdstr)
			if execErr != nil {
				time.Sleep(time.Second * 3)
				continue
			}
			break
		}
		if execErr != nil {
			return execErr
		}

		for retry := 0; retry < 3; retry++ {
			time.Sleep(time.Second * 2)
			result, err := c.LXC.GetCMDResult(execId)
			execErr = err
			if execErr != nil {
				continue
			}

			switch result.Status {
			case "FAILED":
				if result.Error != nil {
					execErr = errors.New(*result.Error)
				}
				continue
			case "RUNNING":
				tflog.Info(ctx, "cmd still running", map[string]any{"try": retry + 1, "cmd": cmdstr})

				// just to stop incrementing
				retry = retry - 1
				continue
			case "SUCCEEDED":
				if *result.ExitCode != 0 {
					execErr = errors.New(*result.Output)
				} else {
					tflog.Info(ctx, "cmd succeeded", map[string]any{"try": retry + 1, "cmd": cmdstr})
				}
				break
			}

		}
		if execErr != nil {
			return execErr
		}
	}

	return nil
}

func deleteLXC(
	ctx context.Context,
	c *pve.PVE,
	node string,
	vmid int,
) error {
	time.Sleep(time.Second * 15)
	// Stop the lxc if running
	if err := updateLXCStatus(
		ctx,
		c,
		node,
		vmid,
		string(pve.LXC_STATUS_STOPPED),
	); err != nil {
		return err
	}

	// check if stopped
	for true {
		time.Sleep(time.Second * 5)
		remoteStatus, err := c.LXC.GetStatus(node, vmid)
		if err != nil {
			return err
		}
		if remoteStatus.Status != string(pve.LXC_STATUS_STOPPED) {
			continue
		}
		break
	}

	if _, err := c.LXC.Delete(node, vmid, nil); err != nil {
		return err
	}

	// check if deleted
	for true {
		time.Sleep(time.Second * 2)
		idAvailable, err := c.Cluster.IsVMIDAvailable(vmid)
		if err != nil {
			return err
		}
		if !idAvailable {
			continue
		}
		break
	}
	return nil
}

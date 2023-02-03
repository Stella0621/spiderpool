// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

package ippoolmanager

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/spidernet-io/spiderpool/pkg/constant"
	spiderpoolip "github.com/spidernet-io/spiderpool/pkg/ip"
	spiderpoolv1 "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v1"
	"github.com/spidernet-io/spiderpool/pkg/logutils"
	"github.com/spidernet-io/spiderpool/pkg/types"
)

func (iw *IPPoolWebhook) mutateIPPool(ctx context.Context, ipPool *spiderpoolv1.SpiderIPPool) error {
	logger := logutils.FromContext(ctx)
	logger.Info("Start to mutate IPPool")

	if ipPool.DeletionTimestamp != nil {
		logger.Info("Terminating IPPool, noting to mutate")
		return nil
	}

	if iw.EnableSpiderSubnet {
		if err := iw.setControllerSubnet(ctx, ipPool); err != nil {
			return apierrors.NewInternalError(fmt.Errorf("failed to set the reference of the controller Subnet: %v", err))
		}
	}

	if !controllerutil.ContainsFinalizer(ipPool, constant.SpiderFinalizer) {
		controllerutil.AddFinalizer(ipPool, constant.SpiderFinalizer)
		logger.Sugar().Infof("Add finalizer %s", constant.SpiderFinalizer)
	}

	if ipPool.Spec.IPVersion == nil {
		var version types.IPVersion
		if spiderpoolip.IsIPv4CIDR(ipPool.Spec.Subnet) {
			version = constant.IPv4
		} else if spiderpoolip.IsIPv6CIDR(ipPool.Spec.Subnet) {
			version = constant.IPv6
		} else {
			return errors.New("invalid 'spec.ipVersion', noting to mutate")
		}

		ipPool.Spec.IPVersion = new(types.IPVersion)
		*ipPool.Spec.IPVersion = version
		logger.Sugar().Infof("Set 'spec.ipVersion' to %d", version)
	}

	if len(ipPool.Spec.IPs) > 1 {
		mergedIPs, err := spiderpoolip.MergeIPRanges(*ipPool.Spec.IPVersion, ipPool.Spec.IPs)
		if err != nil {
			return fmt.Errorf("failed to merge 'spec.ips': %v", err)
		}

		ipPool.Spec.IPs = mergedIPs
		logger.Sugar().Debugf("Merge 'spec.ips':\n%v\n\nto:\n\n%v", ipPool.Spec.IPs, mergedIPs)
	}

	if len(ipPool.Spec.ExcludeIPs) > 1 {
		mergedExcludeIPs, err := spiderpoolip.MergeIPRanges(*ipPool.Spec.IPVersion, ipPool.Spec.ExcludeIPs)
		if err != nil {
			return fmt.Errorf("failed to merge 'spec.excludeIPs': %v", err)
		}

		ipPool.Spec.ExcludeIPs = mergedExcludeIPs
		logger.Sugar().Debugf("Merge 'spec.excludeIPs':\n%v\n\nto:\n\n%v", ipPool.Spec.ExcludeIPs, mergedExcludeIPs)
	}

	return nil
}

func (iw *IPPoolWebhook) setControllerSubnet(ctx context.Context, ipPool *spiderpoolv1.SpiderIPPool) error {
	logger := logutils.FromContext(ctx)

	// TODO(iiiceoo): There was an occasional bug.
	owner := metav1.GetControllerOf(ipPool)
	if v, ok := ipPool.Labels[constant.LabelIPPoolOwnerSpiderSubnet]; ok && owner != nil && v == owner.Name {
		return nil
	}

	var subnetList spiderpoolv1.SpiderSubnetList
	if err := iw.Client.List(ctx, &subnetList); err != nil {
		return fmt.Errorf("failed to list Subnets: %v", err)
	}

	for _, subnet := range subnetList.Items {
		if subnet.Spec.Subnet == ipPool.Spec.Subnet {
			if !metav1.IsControlledBy(ipPool, &subnet) {
				if err := ctrl.SetControllerReference(&subnet, ipPool, iw.Scheme); err != nil {
					return fmt.Errorf("failed to set owner reference: %v", err)
				}
				logger.Sugar().Infof("Set owner reference as Subnet %s", subnet.Name)
			}

			if ipPool.Labels == nil {
				ipPool.Labels = make(map[string]string)
			}
			if v, ok := ipPool.Labels[constant.LabelIPPoolOwnerSpiderSubnet]; !ok || v != subnet.Name {
				ipPool.Labels[constant.LabelIPPoolOwnerSpiderSubnet] = subnet.Name
				logger.Sugar().Infof("Set label %s: %s", constant.LabelIPPoolOwnerSpiderSubnet, subnet.Name)
			}
			break
		}
	}

	return nil
}

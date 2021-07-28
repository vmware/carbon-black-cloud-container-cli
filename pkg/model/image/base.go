/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package image

import (
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/vmware/carbon-black-cloud-container-cli/internal/config"
)

const k8sOverviewTemplate = "/kubernetes/image/%s/overview"

// Identifier is the identifier for an image.
type Identifier struct {
	FullTag        string   `json:"full_tag"`
	Registry       string   `json:"registry"`
	Repo           string   `json:"repo"`
	Tag            string   `json:"tag"`
	ManifestDigest string   `json:"manifest_digest"`
	RepoDigests    []string `json:"repo_digests"`
}

// Metadata is the metadata of the image.
type Metadata struct {
	Distro        string `json:"distro"`
	DistroVersion string `json:"distro_version"`
	ImageSize     uint   `json:"image_size"`
	LayerCount    uint   `json:"layer_count"`
}

// Footer for Identifier will provide the overview link.
func (i *Identifier) Footer() string {
	reportLink, err := url.Parse(config.GetConfig(config.SaasURL))
	if err != nil {
		logrus.Fatal(fmt.Errorf("failed to parse SaaS URL: %w", err))
	}

	reportLink.Path = fmt.Sprintf(k8sOverviewTemplate, i.ManifestDigest)

	footer := fmt.Sprintf("Detailed report can be found at\n%s", reportLink)

	return footer
}

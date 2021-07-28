/*
 * Copyright 2021 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package scan

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anchore/syft/syft/source"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/vmware/carbon-black-cloud-container-cli/pkg/model/bom"
)

func TestGenerateBomOk(t *testing.T) {
	registryHandler := RegistryHandler{registryImageCopy: mockCopyImage}

	bom, err := registryHandler.Generate("foo:latest", Option{})
	if err != nil {
		t.Errorf("failed to generate BOM: %v", err)
	}

	if bom == nil {
		t.Error("BOM was unexpectedly nil")
	}
}

func TestGenerateBomBadImage(t *testing.T) {
	registryHandler := RegistryHandler{registryImageCopy: mockCopyImage}

	// Unknown scheme
	bom, err := registryHandler.Generate("foo://latest", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// No such file
	bom, err = registryHandler.Generate("/foo/bar/baz", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// Unable to expand home directory
	bom, err = registryHandler.Generate("~foo", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}

	// Unknown source
	bom, err = registryHandler.Generate("foo:bar:baz", Option{})
	if err == nil {
		if bom == nil {
			t.Error("Missing expected error while parsing SBOM")
		} else {
			t.Errorf("Expected error but received SBOM: %v", bom)
		}
	}
}

func TestAttachTag(t *testing.T) {
	registry := "registry.com"
	repo := "repo"
	tag := "tag"
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte("digest")))

	// the doc already contains one valid tag
	input := fmt.Sprintf("%s/%s:%s-input", registry, repo, tag)
	fullTag := fmt.Sprintf("%s/%s:%s", registry, repo, tag)
	doc := bom.JSONDocument{
		Source: bom.JSONSource{
			Target: bom.JSONImageSource{
				ImageMetadata: source.ImageMetadata{
					Tags: []string{fullTag},
				},
			},
		},
	}

	returnTag := attachTag(&doc, input, Option{})
	if returnTag != fullTag {
		t.Errorf("Unexpected full tag returned: %s", returnTag)
	}

	// the input is "repo@sha256:xxx"
	input = fmt.Sprintf("%s@sha256:%s", repo, digest)
	fullTag = fmt.Sprintf("docker.io/library/%s@sha256:%s", repo, digest)
	doc = bom.JSONDocument{
		Source: bom.JSONSource{
			Target: bom.JSONImageSource{
				ImageMetadata: source.ImageMetadata{
					Tags: []string{},
				},
			},
		},
	}

	returnTag = attachTag(&doc, input, Option{})
	if returnTag != fullTag {
		t.Errorf("Unexpected full tag returned: %s", returnTag)
	}

	if doc.Source.Target.(bom.JSONImageSource).Tags[0] !=
		fmt.Sprintf("docker.io/library/%s:sha256_%s", repo, digest) {
		t.Errorf("Unexpected full tag attached in doc: %+v", doc)
	}

	// the input is "registry/repo@sha256:xxx"
	input = fmt.Sprintf("%s/%s@sha256:%s", registry, repo, digest)
	doc = bom.JSONDocument{
		Source: bom.JSONSource{
			Target: bom.JSONImageSource{
				ImageMetadata: source.ImageMetadata{
					Tags: []string{},
				},
			},
		},
	}

	returnTag = attachTag(&doc, input, Option{})
	if returnTag != input {
		t.Errorf("Unexpected full tag returned: %s", returnTag)
	}

	if doc.Source.Target.(bom.JSONImageSource).Tags[0] !=
		fmt.Sprintf("%s/%s:sha256_%s", registry, repo, digest) {
		t.Errorf("Unexpected full tag attached in doc: %+v", doc)
	}

	// the input is a tar bar
	input = fmt.Sprintf("%s.tar", repo)
	doc = bom.JSONDocument{
		Source: bom.JSONSource{
			Target: bom.JSONImageSource{
				ImageMetadata: source.ImageMetadata{
					Tags:           []string{},
					ManifestDigest: digest,
				},
			},
		},
	}

	returnTag = attachTag(&doc, input, Option{})
	if returnTag != fmt.Sprintf("%s:%s", repo, digest) {
		t.Errorf("Unexpected full tag returned: %s", returnTag)
	}

	if doc.Source.Target.(bom.JSONImageSource).Tags[0] !=
		fmt.Sprintf("%s:%s", repo, digest) {
		t.Errorf("Unexpected full tag attached in doc: %+v", doc)
	}
}

func mockCopyImage(
	ctx context.Context,
	policyContext *signature.PolicyContext,
	destRef,
	srcRef types.ImageReference,
	options *copy.Options,
) (copiedManifest []byte, retErr error) {
	// fetch the test file position and split out the base dir
	_, testPath, _, _ := runtime.Caller(0) // nolint:dogsled
	baseDir := strings.Split(testPath, "pkg")[0]

	// get the position of the test tar file
	testFixture := filepath.Join(baseDir, "test/tarfile/alpine_3.13.5.tar")

	src, err := alltransports.ParseImageName(fmt.Sprintf("docker-archive:%s", testFixture))
	if err != nil {
		return nil, err
	}

	return copy.Image(ctx, policyContext, destRef, src, options)
}

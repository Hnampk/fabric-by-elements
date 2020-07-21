/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package packaging_test

import (
	"testing"

	"example.com/custom-sdk/fabric/usable-inter-nal/peer/packaging"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate counterfeiter -o mock/platform.go --fake-name Platform . platform
type platform interface {
	packaging.Platform
}

func TestPackaging(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Platforms Suite")
}

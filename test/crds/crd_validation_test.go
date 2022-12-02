//go:build unit

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/3scale/apicast-operator/apis/apps/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/validation"
	"github.com/ghodss/yaml"

	"github.com/stretchr/testify/assert"
)

type testCRInfo struct {
	crPrefix   string
	apiVersion string
}

func TestSampleCustomResources(t *testing.T) {
	schemaRoot := "../../bundle/manifests"
	samplesRootList := []string{
		"../../config/samples",
		"../../doc/cr_samples",
	}
	crdCrMap := map[string]testCRInfo{
		"apps.3scale.net_apicasts.yaml": testCRInfo{
			crPrefix:   "apps_v1alpha1_apicast",
			apiVersion: v1alpha1.GroupVersion.Version,
		},
	}
	for crd, elem := range crdCrMap {
		for _, samplesRoot := range samplesRootList {
			validateCustomResources(t, schemaRoot, samplesRoot, crd, elem.crPrefix, elem.apiVersion)
		}
	}
}

func validateCustomResources(t *testing.T, schemaRoot, samplesRoot, crd, prefix string, version string) {
	schema := getSchemaVersioned(t, fmt.Sprintf("%s/%s", schemaRoot, crd), version)
	assert.NotNil(t, schema)
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), prefix) {
			bytes, err := ioutil.ReadFile(path)
			assert.NoError(t, err, "Error reading CR yaml from %v", path)
			var input map[string]interface{}
			assert.NoError(t, yaml.Unmarshal(bytes, &input))
			assert.NoError(t, schema.Validate(input), "File %v does not validate against the %s CRD schema", info.Name(), crd)
		}
		return nil
	}
	err := filepath.Walk(samplesRoot, walkFunc)
	assert.NoError(t, err, "Error reading CR yaml files from ", samplesRoot)
}

type testCRDInfo struct {
	obj        interface{}
	apiVersion string
}

func TestCompleteCRD(t *testing.T) {
	root := "../../bundle/manifests"

	crdStructMap := map[string]testCRDInfo{
		"apps.3scale.net_apicasts.yaml": testCRDInfo{
			obj:        &v1alpha1.APIcast{},
			apiVersion: v1alpha1.GroupVersion.Version,
		},
	}
	for crd, elem := range crdStructMap {
		schema := getSchemaVersioned(t, fmt.Sprintf("%s/%s", root, crd), elem.apiVersion)
		missingEntries := schema.GetMissingEntries(elem.obj)
		for _, missing := range missingEntries {
			assert.Fail(t, "Discrepancy between CRD and Struct", "CRD: %s: Missing or incorrect schema validation at %s, expected type %s", crd, missing.Path, missing.Type)
		}
	}
}

func getSchemaVersioned(t *testing.T, crd string, version string) validation.Schema {
	bytes, err := ioutil.ReadFile(crd)
	assert.NoError(t, err, "Error reading CRD yaml from %v", crd)
	schema, err := validation.NewVersioned(bytes, version)
	assert.NoError(t, err)
	return schema
}

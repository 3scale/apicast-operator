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

func TestSampleCustomResources(t *testing.T) {
	schemaRoot := "../../bundle/manifests"
	samplesRoot := "../../config/samples"
	crdCrMap := map[string]string{
		"apps.3scale.net_apicasts.yaml": "apps_v1alpha1_apicast",
	}
	for crd, prefix := range crdCrMap {
		validateCustomResources(t, schemaRoot, samplesRoot, crd, prefix)
	}
}

func validateCustomResources(t *testing.T, schemaRoot, samplesRoot, crd, prefix string) {
	schema := getSchema(t, fmt.Sprintf("%s/%s", schemaRoot, crd))
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

func TestCompleteCRD(t *testing.T) {
	root := "../../bundle/manifests"
	crdStructMap := map[string]interface{}{
		"apps.3scale.net_apicasts.yaml": &v1alpha1.APIcast{},
	}
	for crd, obj := range crdStructMap {
		schema := getSchema(t, fmt.Sprintf("%s/%s", root, crd))
		missingEntries := schema.GetMissingEntries(obj)
		for _, missing := range missingEntries {
			assert.Fail(t, "Discrepancy between CRD and Struct", "CRD: %s: Missing or incorrect schema validation at %s, expected type %s", crd, missing.Path, missing.Type)
		}
	}
}

func getSchema(t *testing.T, crd string) validation.Schema {
	bytes, err := ioutil.ReadFile(crd)
	assert.NoError(t, err, "Error reading CRD yaml from %v", crd)
	schema, err := validation.New(bytes)
	assert.NoError(t, err)
	return schema
}

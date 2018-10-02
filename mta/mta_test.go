package mta

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cloud-mta-build-tool/cmd/logs"
)

type testInfo struct {
	name      string
	expected  Modules
	validator func(t *testing.T, actual, expected Modules)
}

func doTest(t *testing.T, expected []testInfo, filename string) {
	logs.NewLogger()
	mtaFile, _ := ioutil.ReadFile(filename)

	actual, _ := Parse(mtaFile)
	for i, tt := range expected {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, actual)
			require.Len(t, actual.Modules, len(expected))
			tt.validator(t, *actual.Modules[i], tt.expected)
		})
	}
	mtaContent, err := Marshal(actual)
	assert.Nil(t, err)
	newActual, newErr := Parse(mtaContent)
	assert.Nil(t, newErr)
	assert.Equal(t, actual, newActual)
}

// Table driven test
// Unit test for parsing mta files to working object
func Test_ModulesParsing(t *testing.T) {
	tests := []testInfo{
		{
			name: "Parse service(srv) Module section",
			expected: Modules{
				Name: "srv",
				Type: "java",
				Path: "srv",
				Requires: []Requires{
					{
						Name: "db",
						Properties: Properties{
							"JBP_CONFIG_RESOURCE_CONFIGURATION": `[tomcat/webapps/ROOT/META-INF/context.xml: {"service_name_for_DefaultDB" : "~{hdi-container-name}"}]`,
						},
					},
				},
				Provides: []Provides{
					{
						Name: "srv_api",
						Properties: Properties{
							"url": "${default-url}",
						},
					},
				},
				Parameters: Parameters{
					"memory": "512M",
				},
				Properties: Properties{
					"APPC_LOG_LEVEL":              "info",
					"VSCODE_JAVA_DEBUG_LOG_LEVEL": "ALL",
				},
			},
			validator: func(t *testing.T, actual, expected Modules) {
				assert.Equal(t, expected.Name, actual.Name)
				assert.Equal(t, expected.Type, actual.Type)
				assert.Equal(t, expected.Path, actual.Path)
				assert.Equal(t, expected.Parameters, actual.Parameters)
				assert.Equal(t, expected.Properties, actual.Properties)
				assert.Equal(t, expected.Requires, actual.Requires)
				assert.Equal(t, expected.Provides, actual.Provides)
			}},

		// ------------------------Second module test------------------------------
		{
			name: "Parse UI(HTML5) Module section",
			expected: Modules{
				Name: "ui",
				Type: "html5",
				Path: "ui",
				Requires: []Requires{
					{
						Name:  "srv_api",
						Group: "destinations",
						Properties: Properties{
							"forwardAuthToken": "true",
							"strictSSL":        "false",
							"name":             "srv_api",
							"url":              "~{url}",
						},
					},
				},
				Provides: []Provides{
					{
						Name: "srv_api",
						Properties: Properties{
							"url": "${default-url}",
						},
					},
				},
				BuildParams: BuildParameters{
					Builder: "grunt",
				},

				Parameters: Parameters{
					"disk-quota": "256M",
					"memory":     "256M",
				},
			},
			validator: func(t *testing.T, actual, expected Modules) {
				assert.Equal(t, expected.Name, actual.Name)
				assert.Equal(t, expected.Type, actual.Type)
				assert.Equal(t, expected.Path, actual.Path)
				assert.Equal(t, expected.Requires, actual.Requires)
				assert.Equal(t, expected.Parameters, actual.Parameters)
				assert.Equal(t, expected.BuildParams, actual.BuildParams)
			}},
	}

	doTest(t, tests, "./testdata/mta.yaml")

}

func Test_BrokenMta(t *testing.T) {
	mtaContent, _ := ioutil.ReadFile("./testdata/mtaWithBrokenProperties.yaml")

	mta, err := Parse(mtaContent)
	require.NotNil(t, err)
	require.NotNil(t, mta)
}

func Test_FullMta(t *testing.T) {
	schemaVersion := "2.0.0"

	expected := MTA{
		SchemaVersion: &schemaVersion,
		Id:            "cloud.samples.someproj",
		Version:       "1.0.0",
		Parameters: Parameters{
			"deploy_mode": "html5-repo",
		},
		Modules: []*Modules{
			{
				Name: "someproj-db",
				Type: "hdb",
				Path: "db",
				Requires: []Requires{
					{
						Name: "someproj-hdi-container",
					},
					{
						Name: "someproj-logging",
					},
				},
				Parameters: Parameters{
					"disk-quota": "256M",
					"memory":     "256M",
				},
			},
			{
				Name: "someproj-java",
				Type: "java",
				Path: "srv",
				Parameters: Parameters{
					"memory":     "512M",
					"disk-quota": "256M",
				},
				Provides: []Provides{
					{
						Name: "java",
						Properties: Properties{
							"url": "${default-url}",
						},
					},
				},
				Requires: []Requires{
					{
						Name: "someproj-hdi-container",
						Properties: Properties{
							"JBP_CONFIG_RESOURCE_CONFIGURATION": "[tomcat/webapps/ROOT/META-INF/context.xml: " +
								"{\"service_name_for_DefaultDB\" : \"~{hdi-container-name}\"}]",
						},
					},
					{
						Name: "someproj-logging",
					},
				},
				BuildParams: BuildParameters{
					Requires: []BuildRequires{
						{
							Name:       "someproj-db",
							TargetPath: "",
						},
					},
				},
			},
			{
				Name: "someproj-catalog-ui",
				Type: "html5",
				Path: "someproj-someprojCatalog",
				Parameters: Parameters{
					"memory":     "256M",
					"disk-quota": "256M",
				},
				Requires: []Requires{
					{
						Name:  "java",
						Group: "destinations",
						Properties: Properties{
							"name": "someproj-backend",
							"url":  "~{url}",
						},
					},
					{
						Name: "someproj-logging",
					},
				},
				BuildParams: BuildParameters{
					Builder: "grunt",
					Requires: []BuildRequires{
						{
							Name:       "someproj-java",
							TargetPath: "",
						},
					},
				},
			},
			{
				Name: "someproj-uideployer",
				Type: "com.sap.html5.application-content",
				Parameters: Parameters{
					"memory":     "256M",
					"disk-quota": "256M",
				},
				Requires: []Requires{
					{
						Name: "someproj-apprepo-dt",
					},
				},
				BuildParams: BuildParameters{
					Builder: "grunt",
					Type:    "com.sap.html5.application-content",
					Requires: []BuildRequires{
						{
							Name: "someproj-catalog-ui",
						},
					},
				},
			},
			{
				Name: "someproj",
				Type: "approuter.nodejs",
				Path: "approuter",
				Parameters: Parameters{
					"memory":     "256M",
					"disk-quota": "256M",
				},
				Requires: []Requires{
					{
						Name:  "java",
						Group: "destinations",
						Properties: Properties{
							"name": "someproj-backend",
							"url":  "~{url}",
						},
					},
					{
						Name: "someproj-apprepo-rt",
					},
					{
						Name: "someproj-logging",
					},
				},
			},
		},
		Resources: []*Resources{
			{
				Name: "someproj-hdi-container",
				Properties: Properties{
					"hdi-container-name": "${service-name}",
				},
				Type: "com.sap.xs.hdi-container",
			},
			{
				Name: "someproj-apprepo-rt",
				Type: "org.cloudfoundry.managed-service",
				Parameters: Parameters{
					"service":      "html5-apps-repo",
					"service-plan": "app-runtime",
				},
			},
			{
				Name: "someproj-apprepo-dt",
				Type: "org.cloudfoundry.managed-service",
				Parameters: Parameters{
					"service":      "html5-apps-repo",
					"service-plan": "app-host",
				},
			},
			{
				Name: "someproj-logging",
				Type: "org.cloudfoundry.managed-service",
				Parameters: Parameters{
					"service":      "application-logs",
					"service-plan": "lite",
				},
			},
		},
	}

	mtaContent, _ := ioutil.ReadFile("./testdata/mta2.yaml")

	actual, err := Parse(mtaContent)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

}
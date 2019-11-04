package scripts

import (
	"sort"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NonManadatoryFields -
type NonManadatoryFields struct {
	SwaggerPath string
}

// ParseSchema -
func ParseSchema(swaggerPath string, logger *logrus.Logger) (*NonManadatoryFields, error) {
	doc, err := loads.Spec(swaggerPath)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err":         err,
			"swaggerPath": swaggerPath,
		}).Error("ParseSchema: loads.Spec(swaggerPath)")
		return nil, errors.Wrapf(err, "ParseSchema: loads.Spec(path), swaggerPath=%+v", swaggerPath)
	}

	expanded, err := doc.Expanded(nil)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err":         err,
			"doc":         doc,
			"swaggerPath": swaggerPath,
		}).Error("ParseSchema: doc.Expanded(nil)")
		return nil, errors.Wrapf(err, "ParseSchema: doc.Expanded(nil), swaggerPath=%+v", swaggerPath)
	}

	nonManadatoryFields := &NonManadatoryFields{
		SwaggerPath: swaggerPath,
	}
	spec := expanded.Spec()
	// data, err := json.Marshal(spec.Definitions["CurrencyAndAmount"])
	// logger.Infof("Definitions=", string(data))

	// testDefinition := spec.Definitions["CurrencyAndAmount"]
	// logger.Infof(
	// 	"spec.Definitions[\"CurrencyAndAmount\"].SchemaProps.Required=%#v",
	// 	testDefinition.SchemaProps.Required,
	// )par

	for path, props := range spec.Paths.Paths {
		// if path != `/accounts/{AccountId}/transactions` {
		// 	continue
		// }

		if props.Delete != nil {
			logger.Infof("Path=%s Delete\n", path)
			PrintResponses(props.Delete, logger)
		}
		if props.Get != nil {
			logger.Infof("Path=%s Get\n", path)
			// data, err := json.Marshal(props.Get)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// logger.Infof("spec.Paths.Paths=", string(data))
			PrintResponses(props.Get, logger)
		}
		if props.Head != nil {
			logger.Infof("Path=%s Head\n", path)
			PrintResponses(props.Head, logger)
		}
		if props.Options != nil {
			logger.Infof("Path=%s Options\n", path)
			PrintResponses(props.Options, logger)
		}
		if props.Patch != nil {
			logger.Infof("Path=%s Patch\n", path)
			PrintResponses(props.Patch, logger)
		}
		if props.Post != nil {
			logger.Infof("Path=%s Post\n", path)
			// data, err := json.Marshal(props.Post)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// fmt.Println("spec.Paths.Paths=", string(data))
			PrintResponses(props.Post, logger)
		}
		if props.Put != nil {
			logger.Infof("Path=%s Put\n", path)
			PrintResponses(props.Put, logger)
		}
	}

	return nonManadatoryFields, nil
}

// PrintResponses -
func PrintResponses(operation *spec.Operation, logger *logrus.Logger) {
	// logger.Infof("Default=%#v\n", operation.Responses.Default)

	for statusCode, response := range operation.Responses.StatusCodeResponses {
		// if statusCode != 201 {
		// 	continue
		// }

		// logger.Infof("statusCode=%d\n", statusCode)
		// logger.Infof("Required=%#v\n", response.Schema.Required)
		// .ResponseProps.Schema.Required
		// fmt.Println()

		if response.ResponseProps.Schema == nil {
			// logger.Infof("statusCode=%d, response.ResponseProps.Schema=nil\n", statusCode)
			continue
		}

		optionalProperties := []string{}
		for propName, prop := range response.ResponseProps.Schema.Properties {
			opts := GetOptionalProperties(propName, prop, "\t", []string{})
			optionalProperties = append(optionalProperties, opts...)
		}

		logger.Infof("\tstatusCode=%d\n", statusCode)
		for _, optionalProperty := range optionalProperties {
			logger.Infof("\t%#v\n", optionalProperty)
		}
		logger.Infof("\n")
	}
}

// GetOptionalProperties -
func GetOptionalProperties(propName string, prop spec.Schema, indent string, path []string) []string {
	// logger.Infof("%sName=%#v\n", indent, propName)
	// logger.Infof("%sDescription=%#v\n", indent, prop.Description)
	// logger.Infof("%sID=%#v\n", indent, prop.Schema)
	// logger.Infof("%spath=%#v\n", indent, path)
	// logger.Infof("%sType=%#v\n", indent, prop.Type)

	newPath := path
	if prop.Type.Contains("array") {
		newPath = append(path, propName+"[*]")
	} else {
		newPath = append(path, propName)
	}

	if (prop.Items == nil || prop.Items.Schema == nil) && len(prop.Properties) <= 0 {
		return []string{strings.Join(newPath, ".")}
	}

	optionalProperties := []string{}
	if prop.Items != nil && prop.Items.Schema != nil {
		required := prop.Items.Schema.Required
		if len(required) > 0 {
			sort.Strings(required)
		}
		// logger.Infof("%sProperties=%#v\n", indent, len(prop.Items.Schema.Properties))
		// logger.Infof("%sRequired=%#v\n", indent, required)

		for propName, prop := range prop.Items.Schema.Properties {
			pos := sort.SearchStrings(required, propName)
			found := pos < len(required) && required[pos] == propName
			if !found {
				opts := GetOptionalProperties(propName, prop, indent+"\t", newPath)
				optionalProperties = append(optionalProperties, opts...)
			}
		}
	} else {
		required := prop.Required
		if len(required) > 0 {
			sort.Strings(required)
		}
		// logger.Infof("%sProperties=%#v\n", indent, len(prop.Properties))
		// logger.Infof("%sRequired=%#v\n", indent, required)

		for propName, prop := range prop.Properties {
			pos := sort.SearchStrings(required, propName)
			found := pos < len(required) && required[pos] == propName
			if !found {
				opts := GetOptionalProperties(propName, prop, indent+"\t", newPath)
				optionalProperties = append(optionalProperties, opts...)
			}
		}
	}

	return optionalProperties
}

package schema

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// Response represents a response object from a HTTP Call
type Response struct {
	Method     string
	Path       string
	Header     http.Header
	Body       io.Reader
	StatusCode int
}

// Failure represents a validation failure
type Failure struct {
	Message string
}

func newFailure(message string) Failure {
	return Failure{
		Message: message,
	}
}

// Validator validates a HTTP response object against a schema
type Validator interface {
	Validate(Response) ([]Failure, error)
	IsRequestProperty(method, path, propertpath string) (bool, error)
}

func NewSwaggerOBSpecValidator(specName, version string) (Validator, error) {
	var err error

	dirnameIndex := 0
	dirnames := []string{"pkg/schema/spec/v3.1.0", "../../pkg/schema/spec/v3.1.0"}
	files := []os.FileInfo{}
	for index, dirname := range dirnames {
		filesReadDir, errReadDir := ioutil.ReadDir(dirname)
		if errReadDir != nil {
			wd, errGetwd := os.Getwd()
			if errGetwd != nil {
				err = errors.Wrapf(errGetwd, "schema: opening spec folder failed in os.Getwd, dirname=%q", dirname)
			} else {
				err = errors.Wrapf(errReadDir, "schema: opening spec folder failed, dirname=%q, wd=%q", dirname, wd)
			}
		} else {
			err = nil
			dirnameIndex = index
			files = filesReadDir
			break
		}
	}

	if err != nil {
		return nil, err
	}

	dirname := dirnames[dirnameIndex]
	for _, f := range files {
		filename := dirname + "/" + f.Name()
		doc, err := loads.Spec(filename)
		if err != nil {
			return nil, errors.Wrapf(err, "schema: opening spec file, filename=%q", filename)
		}
		if doc.Spec().Info.Version == version && doc.Spec().Info.Title == specName {
			return NewSwaggerValidator(filename)
		}
	}

	return nil, fmt.Errorf("schema: could not find spec file for spec %s version %s", specName, version)
}

// NewSwaggerValidator returns a swagger validator implementation
// Takes a schema file path as source, can be remote http(s) or local
func NewSwaggerValidator(schemaPath string) (Validator, error) {
	doc, err := loads.Spec(schemaPath)
	if err != nil {
		return nil, err
	}
	return newValidator(doc)
}

type validators struct {
	validators []Validator
	document   *loads.Document
}

func newValidator(doc *loads.Document) (Validator, error) {
	f := newFinder(doc)

	if doc.Version() != "2.0" {
		return nil, errors.New("unsupported swagger version")
	}

	specVersion := doc.Spec().Info.Version
	switch specVersion {
	case "v3.0.0":
		fallthrough
	case "v3.1.0":
		return validators{
			validators: []Validator{
				newContentTypeValidator(f),
				newStatusCodeValidator(f),
				newBodyValidator(f),
			},
			document: doc,
		}, nil
	}

	return nil, errors.New("unsupported spec version")
}

func (v validators) Validate(r Response) ([]Failure, error) {
	allFailures := []Failure{}
	for _, validator := range v.validators {
		failures, err := validator.Validate(r)
		if err != nil {
			return nil, err
		}
		allFailures = append(allFailures, failures...)
	}
	return allFailures, nil
}

func (v validators) IsRequestProperty(checkmethod, checkpath, propertyPath string) (bool, error) {
	spec := v.document.Spec()

	for path, props := range spec.Paths.Paths {
		for method, op := range getOperations(&props) {
			if path == checkpath && method == checkmethod {
				for _, param := range op.Parameters {
					if param.ParamProps.In == "body" {
						schema := param.ParamProps.Schema
						found := findPropertyInSchema(schema, propertyPath, "")
						if found {
							return true, nil
						}
					}
				}
			}
		}
	}

	return false, nil
}

func findPropertyInSchema(sc *spec.Schema, propertyPath, previousPath string) bool {
	for k, j := range sc.SchemaProps.Properties {
		var element string
		if len(previousPath) == 0 {
			element = k
		} else {
			element = previousPath + "." + k
		}
		fmt.Printf("%s\n", element)
		if element == propertyPath {
			return true
		}

		if findPropertyInSchema(&j, propertyPath, element) {
			return true
		}
	}
	return false
}

// getOperations returns a mapping of HTTP Verb name to "spec operation name"
func getOperations(props *spec.PathItem) map[string]*spec.Operation {
	ops := map[string]*spec.Operation{
		"DELETE":  props.Delete,
		"GET":     props.Get,
		"HEAD":    props.Head,
		"OPTIONS": props.Options,
		"PATCH":   props.Patch,
		"POST":    props.Post,
		"PUT":     props.Put,
	}

	// Keep those != nil
	for key, op := range ops {
		if op == nil {
			delete(ops, key)
		}
	}
	return ops
}

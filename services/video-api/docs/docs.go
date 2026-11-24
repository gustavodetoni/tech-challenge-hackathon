// Package docs contains the minimum Swagger document for the foundation API.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{"swagger":"2.0","info":{"title":"FIAP Video API","version":"0.1.0","description":"Foundation endpoints for the FIAP X video platform."},"host":"localhost:8080","basePath":"/api/v1","paths":{"/ping":{"get":{"summary":"Check that the API is available","produces":["application/json"],"responses":{"200":{"description":"OK"}}}}}}`

type swaggerDoc struct{}

func (swaggerDoc) ReadDoc() string { return docTemplate }

func init() { swag.Register("swagger", swaggerDoc{}) }

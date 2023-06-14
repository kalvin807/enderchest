// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/image": {
            "put": {
                "description": "Uploads an image to the server",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Upload image",
                "parameters": [
                    {
                        "type": "file",
                        "description": "Image file",
                        "name": "image",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Image metadata",
                        "name": "metadata",
                        "in": "formData"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Image uploaded successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "get form err",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "failed to save metadata",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "Checks if the server is alive",
                "produces": [
                    "application/json"
                ],
                "summary": "Check server status",
                "responses": {
                    "200": {
                        "description": "pong",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Enderchest API",
	Description:      "This is the API server for Enderchest.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

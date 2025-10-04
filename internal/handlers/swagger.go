package handlers

import "net/http"

var swaggerSpec = []byte(`{
  "openapi": "3.0.0",
  "info": {"title": "KPop REST API", "version": "1.0.0"},
  "paths": {
    "/api/login": {"post": {"summary": "Login", "requestBody": {"required": true}, "responses": {"200": {"description": "OK"}}}},
    "/api/logout": {"post": {"summary": "Logout", "responses": {"200": {"description": "OK"}}}},
    "/api/data": {"get": {"summary": "Secret data", "security": [{"bearerAuth": []}], "responses": {"200": {"description": "OK"}}}},
    "/api/users": {"get": {"summary": "List users", "security": [{"bearerAuth": []}], "responses": {"200": {"description": "OK"}}}},
    "/api/idols": {"get": {"summary": "List idols", "security": [{"bearerAuth": []}]}, "post": {"summary": "Create idol", "security": [{"bearerAuth": []}]}},
    "/api/idols/{id}": {"put": {"summary": "Update idol", "security": [{"bearerAuth": []}]}, "delete": {"summary": "Delete idol", "security": [{"bearerAuth": []}]}}
  },
  "components": {"securitySchemes": {"bearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}}}
}`)

func SwaggerSpec(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write(swaggerSpec)
}

func SwaggerUI(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
  <style>body { margin:0; } #swagger-ui { max-width: 100%; }</style>
  </head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>`))
}



SELECT
  {{ .Ident "guid" }},
  {{ .Ident "name" }},
  {{ .Ident "namespace" }},
  {{ .Ident "api_version" }},
  {{ .Ident "annotations" }},
  {{ .Ident "labels" }},
  {{ .Ident "created" }},
  {{ .Ident "created_by" }},
  {{ .Ident "updated" }},
  {{ .Ident "updated_by" }},
  {{ .Ident "description" }},
  {{ .Ident "type" }},
  {{ .Ident "payload" }}
FROM
  {{ .Ident "secret_keeper" }}
WHERE {{ .Ident "namespace" }} = {{ .Arg .Namespace }}
ORDER BY {{ .Ident "updated" }} DESC
;

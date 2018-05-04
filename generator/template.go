package generator

const tpl = `
{{- /*************** header template *****************/}}
{{define "header" -}}
// !!! DO NOT EDIT THIS FILE. It is generated by 'light' tool.
// @light: https://github.com/arstd/light
// Generated from source: {{.Source}}
package {{.Package}}
import (
		"bytes"
		"fmt"
		"github.com/arstd/light/light"
		"github.com/arstd/light/null"
		{{- if .Log }}
			"github.com/arstd/log"
		{{- end}}

		{{- range $path, $short := .Imports}}
			{{$short}} "{{$path}}"
		{{- end}}
)

{{if .VarName}}
func init() { {{.VarName}} = new(Store{{.Name}}) }
{{end}}

type Store{{.Name}} struct{}
{{end}}

{{- /*************** fragment template *****************/}}
{{define "fragment" -}}
{{- if .Fragment.Condition}}
	if {{.Fragment.Condition}} {
{{- end }}
{{- if .Fragment.Statement }}
	{{- if .Fragment.Range }}
		if len({{.Fragment.Range}}) > 0 {
			fmt.Fprintf(&buf, "{{.Fragment.Statement}} ", strings.Repeat(",?", len({{.Fragment.Range}}))[1:])
			for _, v := range {{.Fragment.Range}} {
				args = append(args, v)
			}
		}
	{{- else if .Fragment.Replacers }}
		fmt.Fprintf(&buf, "{{.Fragment.Statement}} "{{range $elem := .Fragment.Replacers}}, {{$elem}}{{end}})
	{{- else }}
		buf.WriteString("{{.Fragment.Statement}} ")
	{{- end }}
	{{- if .Fragment.Variables }}
		args = append(args{{range $elem := .Fragment.Variables}}, {{ParamsVarByNameValue $.Method.Params $elem}}{{end}})
	{{- end }}
{{- else }}
	{{- range $fragment := .Fragment.Fragments }}
		{{template "fragment" (aggregate $.Method $fragment)}}
	{{- end }}
{{- end }}
{{- if .Fragment.Condition}}
	}
{{- end }}
{{end}}


{{- /*************** ddl template *****************/}}
{{define "ddl" -}}
query := buf.String()
{{- if .Store.Log }}
	log.Debug(query)
	log.Debug(args...)
{{- end}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
_, err := exec.ExecContext(ctx, query, args...)
{{- if .Store.Log }}
	if err != nil {
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	}
{{- end}}
return err
{{end}}

{{- /*************** update/delete template *****************/}}
{{define "update" -}}
query := buf.String()
{{if .Store.Log -}}
	log.Debug(query)
	log.Debug(args...)
{{end -}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
res, err := exec.ExecContext(ctx, query, args...)
if err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return 0, err
}
return res.RowsAffected()
{{end -}}

{{- /*************** insert template *****************/}}
{{define "insert" -}}
query := buf.String()
{{- if .Store.Log }}
	log.Debug(query)
	log.Debug(args...)
{{end}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
res, err := exec.ExecContext(ctx, query, args...)
if err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{end}}
	return 0, err
}
return res.LastInsertId()
{{end}}

{{- /*************** get template *****************/}}
{{define "get" -}}
query := buf.String()
{{- if .Store.Log }}
	log.Debug(query)
	log.Debug(args...)
{{- end}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
row := exec.QueryRowContext(ctx, query, args...)
xu := new({{VariableTypeName .Results.Result}})
xdst := []interface{}{
	{{- range $i, $field := .Statement.Fields -}}
		{{- if $i -}} , {{- end -}}
		{{- call $.ResultVarByTagScan $field -}}
	{{- end -}}
}
err := row.Scan(xdst...)
if err != nil {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	{{- if .Store.Log}}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end }}
		return nil, err
	}
{{- if .Store.Log}}
	log.Trace(xdst)
{{- end }}
return xu, err
{{end}}

{{- /*************** list template *****************/}}
{{define "list" -}}
query := buf.String()
{{- if .Store.Log }}
	log.Debug(query)
	log.Debug(args...)
{{- end}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
rows, err := exec.QueryContext(ctx, query, args...)
if err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return nil, err
}
defer rows.Close()
var data {{VariableTypeName .Results.Result}}
for rows.Next() {
	xu := new({{ VariableElemTypeName .Results.Result }})
	data = append(data, xu)
	xdst := []interface{}{
		{{- range $i, $field := .Statement.Fields -}}
			{{- if $i -}} , {{- end -}}
			{{- call $.ResultVarByTagScan $field -}}
		{{- end -}}
	}
	err = rows.Scan(xdst...)
	if err != nil {
		{{- if .Store.Log }}
			log.Error(query)
			log.Error(args...)
			log.Error(err)
		{{- end}}
		return nil, err
	}
	{{- if .Store.Log }}
		log.Trace(xdst)
	{{- end}}
}
if err = rows.Err(); err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return nil, err
}
return data, nil
{{end}}

{{- /*************** page template *****************/}}
{{define "page" -}}
var total int64
totalQuery := "SELECT count(1) "+ buf.String()
{{- if .Store.Log }}
	log.Debug(totalQuery)
	log.Debug(args...)
{{- end}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
err := exec.QueryRowContext(ctx, totalQuery, args...).Scan(&total)
if err != nil {
	{{- if .Store.Log }}
		log.Error(totalQuery)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return 0, nil, err
}
{{- if .Store.Log }}
	log.Debug(total)
{{- end}}

{{$i := sub (len .Statement.Fragments) 1}}
{{ $fragment := index .Statement.Fragments $i }}
{{template "fragment" (aggregate $ $fragment)}}
{{ $fragement0 := index .Statement.Fragments 0 }}
query := "{{$fragement0.Statement}} " + buf.String()
{{- if .Store.Log }}
	log.Debug(query)
	log.Debug(args...)
{{- end}}
ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
rows, err := exec.QueryContext(ctx, query, args...)
if err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return 0, nil, err
}
defer rows.Close()
var data {{VariableTypeName .Results.Result}}
for rows.Next() {
	xu := new({{ VariableElemTypeName .Results.Result }})
	data = append(data, xu)
	xdst := []interface{}{
		{{- range $i, $field := .Statement.Fields -}}
			{{- if $i -}} , {{- end -}}
			{{- call $.ResultVarByTagScan $field -}}
		{{- end -}}
	}
	err = rows.Scan(xdst...)
	if err != nil {
		{{- if .Store.Log }}
			log.Error(query)
			log.Error(args...)
			log.Error(err)
		{{- end}}
		return 0, nil, err
	}
	{{- if .Store.Log }}
		log.Trace(xdst)
	{{- end}}
}
if err = rows.Err(); err != nil {
	{{- if .Store.Log }}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return 0, nil, err
}
return total, data, nil
{{end}}


{{- /*************** agg template *****************/}}
{{define "agg" -}}
query := buf.String()
{{- if .Store.Log}}
	log.Debug(query)
	log.Debug(args...)
{{- end}}
var agg {{VariableTypeName .Results.Result}}
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
err := exec.QueryRowContext(ctx, query, args...).Scan({{VariableWrap .Results.Result}}(&agg))
if err != nil {
	if err == sql.ErrNoRows {
		{{- if .Store.Log}}
			log.Debug(agg)
		{{- end}}
		return agg, nil
	}
	{{- if .Store.Log}}
		log.Error(query)
		log.Error(args...)
		log.Error(err)
	{{- end}}
	return agg, err
}
{{- if .Store.Log}}
	log.Debug(agg)
{{- end}}
return agg, nil
{{end}}

{{- /*************** main *****************/ -}}
{{template "header" . -}}
{{range $method := .Methods -}}
	func (*Store{{$.Name}}) {{MethodSignature $method}} {
		{{$tx := MethodTx $method -}}
		var exec = {{if $tx }} light.GetExec({{$tx}}, db) {{else}} db {{end}}
		var buf bytes.Buffer
		var args []interface{}

		{{- range $i, $fragment := .Statement.Fragments }}
			{{/* if type=page, return field statement and ordery by limit statement reserved */}}
			{{$last := sub (len $method.Statement.Fragments) 1 }}
			{{if not (and (eq $method.Type "page") (or (eq $i 0) (eq $i $last)))}}
				{{template "fragment" (aggregate $method $fragment)}}
			{{end}}
		{{- end }}

		{{if eq $method.Type "ddl" -}}
			{{template "ddl" $method}}
		{{else if or (eq $method.Type "update") (eq $method.Type "delete") -}}
			{{template "update" $method}}
		{{else if eq $method.Type "insert"}}
			{{template "insert" $method}}
		{{else if eq $method.Type "get"}}
			{{template "get" $method}}
		{{else if eq $method.Type "list"}}
			{{template "list" $method}}
		{{else if eq $method.Type "page"}}
			{{template "page" $method}}
		{{else if eq $method.Type "agg"}}
			{{template "agg" $method}}
		{{else}}
			panic("unimplemented")
		{{end -}}
	}
{{end}}
`

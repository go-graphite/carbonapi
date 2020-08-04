package parser

import (
	"strings"
	"text/template"
)

type defineStruct struct {
	tpl *template.Template
}

var defineMap = &defineStruct{tpl: template.New("define")}

// Define new template
func Define(name, tmpl string) error {
	return defineMap.define(name, tmpl)
}

func defineCleanUp() {
	defineMap = &defineStruct{tpl: template.New("define")}
}

func (d *defineStruct) define(name, tmpl string) error {
	t := d.tpl.New(name)
	_, err := t.Parse(tmpl)
	return err
}

func (d *defineStruct) expandExpr(exp *expr) (*expr, error) {
	if exp == nil {
		return exp, nil
	}

	var err error

	if exp.etype == EtName || exp.etype == EtFunc {
		t := d.tpl.Lookup(exp.target)
		if t != nil {
			var b strings.Builder
			args := make([]string, len(exp.args))
			for i := 0; i < len(exp.args); i++ {
				args[i] = exp.args[i].ToString()
			}
			kwargs := make(map[string]string)
			for k, v := range exp.namedArgs {
				kwargs[k] = v.ToString()
			}
			data := map[string]interface{}{
				"argString": exp.argString,
				"args":      args,
				"kwargs":    kwargs,
			}
			err = t.Execute(&b, data)
			if err != nil {
				return exp, err
			}
			newExp, _, err := parseExprInner(b.String())
			if err != nil {
				return exp, err
			}
			exp = newExp.(*expr)
		}
	}

	for i := 0; i < len(exp.args); i++ {
		exp.args[i], err = d.expandExpr(exp.args[i])
		if err != nil {
			return exp, err
		}
	}

	for k, v := range exp.namedArgs {
		exp.namedArgs[k], err = d.expandExpr(v)
		if err != nil {
			return exp, err
		}
	}

	return exp, nil
}

func (d *defineStruct) Expand(v Expr) (Expr, error) {
	exp, ok := v.(*expr)
	if !ok {
		return v, nil
	}
	return d.expandExpr(exp)
}

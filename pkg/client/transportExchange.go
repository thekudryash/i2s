package client

import (
	"context"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"

	"github.com/seniorGolang/i2s/pkg/meta"
)

func renderTransportExchange(info *meta.GenerationInfo) (err error) {

	srcFile := NewFile(strings.ToLower(info.ServiceName))
	srcFile.PackageComment("GENERATED BY i2s. DO NOT EDIT.")

	if len(info.Iface.Methods) > 0 {
		srcFile.Type().Op("(")
	}

	ctx, _ := prepareContext(info.SourceFilePath, info.Iface)
	ctx = context.WithValue(ctx, "code", srcFile)

	for _, signature := range info.Iface.Methods {
		srcFile.Add(exchange(ctx, requestStructName(signature), removeContextIfFirst(signature.Args)))
		srcFile.Add(exchange(ctx, responseStructName(signature), removeErrorIfLast(signature.Results))).Line()
	}

	if len(info.Iface.Methods) > 0 {
		srcFile.Op(")")
	}
	return srcFile.Save(path.Join(info.OutputFilePath, strings.ToLower(info.ServiceName), "exchange.go"))
}

func exchange(ctx context.Context, name string, params []types.Variable) Code {

	if len(params) == 0 {
		return Comment("Formal exchange type, please do not delete.").Line().Id(name).Struct()
	}

	return Id(name).StructFunc(func(g *Group) {
		for _, param := range params {
			g.Add(structField(ctx, &param))
		}
	})
}
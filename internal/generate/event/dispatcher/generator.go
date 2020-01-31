package dispatcher

import (
	"bytes"
	"go/format"

	"github.com/dave/jennifer/jen"

	"sagikazarmark.dev/mga/pkg/gentypes"
)

// File provides information for generating event dispatchers.
type File struct {
	gentypes.File

	// EventDispatchers represents event dispatchers to be generated for matching interfaces.
	EventDispatchers []EventDispatcher
}

// EventDispatcher describes the event dispatcher interface.
type EventDispatcher struct {
	Name              string
	DispatcherMethods []EventMethod
}

// Generate generates an event dispatcher.
func Generate(file File) ([]byte, error) {
	code := jen.NewFilePathName(file.Package.Path, file.Package.Name)

	code.HeaderComment("// +build !ignore_autogenerated\n")

	if file.HeaderText != "" {
		code.HeaderComment(file.HeaderText)
	}

	code.HeaderComment("Code generated by mga tool. DO NOT EDIT.")

	code.ImportName("emperror.dev/errors", "errors")

	const eventBusTypeName = "EventBus"
	generateEventBus(code, eventBusTypeName)

	for _, eventDispatcher := range file.EventDispatchers {
		generateEventDispatcher(code, eventDispatcher)
	}

	var buf bytes.Buffer

	err := code.Render(&buf)
	if err != nil {
		return nil, err
	}

	return format.Source(buf.Bytes())
}

func generateEventBus(code *jen.File, eventBusTypeName string) {
	code.Commentf("%s is a generic event bus.", eventBusTypeName)
	code.Type().Id(eventBusTypeName).Interface(
		jen.Comment("Publish sends an event to the underlying message bus."),
		jen.Id("Publish").Params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("event").Interface(),
		).Error(),
	).Line()
}

func generateEventDispatcher(code *jen.File, eventDispatcher EventDispatcher) {
	eventDispatcherTypeName := eventDispatcher.Name + "EventDispatcher"

	const (
		eventBusVarName  = "bus"
		eventBusTypeName = "EventBus"
	)

	code.Commentf("%s dispatches events through the underlying generic event bus.", eventDispatcherTypeName)
	code.Type().Id(eventDispatcherTypeName).Struct(
		jen.Id(eventBusVarName).Id(eventBusTypeName),
	).Line()

	code.Commentf("New%s returns a new %s instance.", eventDispatcherTypeName, eventDispatcherTypeName)
	code.Func().
		Id("New" + eventDispatcherTypeName).
		Params(jen.Id(eventBusVarName).Id(eventBusTypeName)).
		Id(eventDispatcherTypeName).
		Block(
			jen.Return(
				jen.Id(eventDispatcherTypeName).Values(jen.Dict{
					jen.Id(eventBusVarName): jen.Id(eventBusVarName),
				}),
			),
		).
		Line()

	for _, method := range eventDispatcher.DispatcherMethods {
		code.ImportName(method.Event.Package.Path, method.Event.Package.Name)

		var params []jen.Code

		if method.ReceivesContext {
			params = append(params, jen.Id("ctx").Qual("context", "Context"))
		}

		params = append(params, jen.Id("event").Qual(method.Event.Package.Path, method.Event.Name))

		code.Commentf("%s dispatches a(n) %s event.", method.Name, method.Event.Name)
		fn := code.Func().Params(
			jen.Id("d").Id(eventDispatcherTypeName),
		).Id(method.Name).Params(params...)

		if method.ReturnsError {
			fn = fn.Error()
		}

		var block []jen.Code

		if !method.ReceivesContext {
			block = append(block, jen.Id("ctx").Op(":=").Qual("context", "Background").Call())
		}

		if method.ReturnsError {
			block = append(
				block,
				jen.Err().Op(":=").Id("d").Dot(eventBusVarName).Dot("Publish").Call(
					jen.Id("ctx"),
					jen.Id("event"),
				),
				jen.If(
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(jen.Qual("emperror.dev/errors", "WithDetails").Call(
						jen.Qual("emperror.dev/errors", "WithMessage").Call(
							jen.Err(),
							jen.Lit("failed to dispatch event"),
						),
						jen.Lit("event"), jen.Lit(method.Event.Name),
					)),
				),
				jen.Line(),
				jen.Return(jen.Nil()),
			)
		} else {
			block = append(block, jen.Id("_").Op("=").Id("d").Dot(eventBusVarName).Dot("Publish").Call(
				jen.Id("ctx"),
				jen.Id("event"),
			))
		}

		fn.Block(block...).Line()
	}
}
